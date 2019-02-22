// Copyright 2017 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package envoy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/pprof"
	"os"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/hashicorp/go-multierror"
	"github.com/prometheus/client_golang/prometheus"

	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pkg/log"
	"istio.io/istio/pkg/util"
	"istio.io/istio/pkg/version"
)

const (
	metricsNamespace     = "pilot"
	metricsSubsystem     = "discovery"
	metricLabelCacheName = "cache_name"
	metricLabelMethod    = "method"
	metricBuildVersion   = "build_version"
)

var (
	// Save the build version information.
	buildVersion = version.Info.String()

	cacheHitCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "cache_hit",
			Help:      "Count of cache hits for a particular cache within Pilot",
		}, []string{metricLabelCacheName, metricBuildVersion})
	callCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "calls",
			Help:      "Counter of individual method calls in Pilot",
		}, []string{metricLabelMethod, metricBuildVersion})
	errorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "errors",
			Help:      "Counter of errors encountered during a given method call within Pilot",
		}, []string{metricLabelMethod, metricBuildVersion})
	webhookCallCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "webhook_calls",
			Help:      "Counter of individual webhook calls made in Pilot",
		}, []string{metricLabelMethod, metricBuildVersion})
	webhookErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "webhook_errors",
			Help:      "Counter of errors encountered when invoking the webhook endpoint within Pilot",
		}, []string{metricLabelMethod, metricBuildVersion})

	resourceBuckets = []float64{0, 10, 20, 30, 40, 50, 75, 100, 150, 250, 500, 1000, 10000}
	resourceCounter = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "resources",
			Help:      "Histogram of returned resource counts per method by Pilot",
			Buckets:   resourceBuckets,
		}, []string{metricLabelMethod, metricBuildVersion})
)

var (
	// Variables associated with clear cache squashing.
	clearCacheMutex sync.Mutex

	// lastClearCache is the time we last pushed
	lastClearCache time.Time

	// lastClearCacheEvent is the time of the last config event
	lastClearCacheEvent time.Time

	// clearCacheEvents is the counter of 'clearCache' calls
	clearCacheEvents int

	// clearCacheTimerSet is true if we are in squash mode, and a timer is already set
	clearCacheTimerSet bool

	// clearCacheTime is the max time to squash a series of events.
	// The push will happen 1 sec after the last config change, or after 'clearCacheTime'
	// Default value is 1 second, or the value of PILOT_CACHE_SQUASH env
	clearCacheTime = 1

	// Push is used to request a push, when config changes. This is used to
	// avoid adding a circular dependency from v1 to v2 - the method is implemented
	// in the ADS server.
	Push func(fullPush bool, edsUpdates map[string]*model.EndpointShardsByService)

	// BeforePush is called before a push. Like Push, it is an artifact of the split
	// of discovery from v2/discovery and the migration from v1.
	BeforePush func() map[string]*model.EndpointShardsByService

	// DebounceAfter is the delay added to events to wait
	// after a registry/config event for debouncing.
	// This will delay the push by at least this interval, plus
	// the time getting subsequent events. If no change is
	// detected the push will happen, otherwise we'll keep
	// delaying until things settle.
	DebounceAfter time.Duration

	// DebounceMax is the maximum time to wait for events
	// while debouncing. Defaults to 10 seconds. If events keep
	// showing up with no break for this time, we'll trigger a push.
	DebounceMax time.Duration
)

func init() {
	// No longer used. Will be removed in 1.1
	//prometheus.MustRegister(cacheSizeGauge)
	//prometheus.MustRegister(cacheHitCounter)
	//prometheus.MustRegister(cacheMissCounter)
	//prometheus.MustRegister(callCounter)
	//prometheus.MustRegister(errorCounter)
	//prometheus.MustRegister(resourceCounter)

	cacheSquash := os.Getenv("PILOT_CACHE_SQUASH")
	if len(cacheSquash) > 0 {
		t, err := strconv.Atoi(cacheSquash)
		if err == nil {
			clearCacheTime = t
		}
	}

	DebounceAfter = envDuration("PILOT_DEBOUNCE_AFTER", 100*time.Millisecond)
	DebounceMax = envDuration("PILOT_DEBOUNCE_MAX", 10*time.Second)
}

func envDuration(env string, def time.Duration) time.Duration {
	envVal := os.Getenv(env)
	if envVal == "" {
		return def
	}
	d, err := time.ParseDuration(envVal)
	if err != nil {
		log.Warnf("Invalid value %s %s %v", env, envVal, err)
		return def
	}
	return d
}

// DiscoveryService publishes services, clusters, and routes for all proxies
type DiscoveryService struct {
	*model.Environment

	webhookClient   *http.Client
	webhookEndpoint string
	// TODO Profile and optimize cache eviction policy to avoid
	// flushing the entire cache when any route, service, or endpoint
	// changes. An explicit cache expiration policy should be
	// considered with this change to avoid memory exhaustion as the
	// entire cache will no longer be periodically flushed and stale
	// entries can linger in the cache indefinitely.
	sdsCache *discoveryCache

	RestContainer *restful.Container

	// true if a full push is needed after debounce. False if only EDS is required.
	fullPush bool
}

type discoveryCacheStatEntry struct {
	Hit  uint64 `json:"hit"`
	Miss uint64 `json:"miss"`
}

type discoveryCacheStats struct {
	Stats map[string]*discoveryCacheStatEntry `json:"cache_stats"`
}

type discoveryCacheEntry struct {
	data          []byte
	hit           uint64 // atomic
	miss          uint64 // atomic
	resourceCount uint32
}

type discoveryCache struct {
	name     string
	disabled bool
	mu       sync.RWMutex
	cache    map[string]*discoveryCacheEntry
}

func newDiscoveryCache(name string, enabled bool) *discoveryCache {
	return &discoveryCache{
		name:     name,
		disabled: !enabled,
		cache:    make(map[string]*discoveryCacheEntry),
	}
}

func (c *discoveryCache) cachedDiscoveryResponse(key string) ([]byte, uint32, bool) {
	if c.disabled {
		return nil, 0, false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	// Miss - entry.miss is updated in updateCachedDiscoveryResponse
	entry, ok := c.cache[key]
	if !ok || entry.data == nil {
		return nil, 0, false
	}

	// Hit
	atomic.AddUint64(&entry.hit, 1)
	cacheHitCounter.With(c.cacheSizeLabels()).Inc()
	return entry.data, entry.resourceCount, true
}

func (c *discoveryCache) resetStats() {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, v := range c.cache {
		atomic.StoreUint64(&v.hit, 0)
		atomic.StoreUint64(&v.miss, 0)
	}
}

func (c *discoveryCache) stats() map[string]*discoveryCacheStatEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := make(map[string]*discoveryCacheStatEntry, len(c.cache))
	for k, v := range c.cache {
		stats[k] = &discoveryCacheStatEntry{
			Hit:  atomic.LoadUint64(&v.hit),
			Miss: atomic.LoadUint64(&v.miss),
		}
	}
	return stats
}

func (c *discoveryCache) cacheSizeLabels() prometheus.Labels {
	return prometheus.Labels{
		metricLabelCacheName: c.name,
		metricBuildVersion:   buildVersion,
	}
}

type hosts struct {
	Hosts []*host `json:"hosts"`
}

type host struct {
	Address string `json:"ip_address"`
	Port    int    `json:"port"`
	Tags    *tags  `json:"tags,omitempty"`
}

type tags struct {
	AZ     string `json:"az,omitempty"`
	Canary bool   `json:"canary,omitempty"`

	// Weight is an integer in the range [1, 100] or empty
	Weight int `json:"load_balancing_weight,omitempty"`
}

type keyAndService struct {
	Key   string  `json:"service-key"`
	Hosts []*host `json:"hosts"`
}

// Request parameters for discovery services
const (
	ServiceCluster = "service-cluster"
	ServiceNode    = "service-node"
)

// DiscoveryServiceOptions contains options for create a new discovery
// service instance.
type DiscoveryServiceOptions struct {
	// The listening address for HTTP. If the port in the address is empty or "0" (as in "127.0.0.1:" or "[::1]:0")
	// a port number is automatically chosen.
	HTTPAddr string

	// The listening address for GRPC. If the port in the address is empty or "0" (as in "127.0.0.1:" or "[::1]:0")
	// a port number is automatically chosen.
	GrpcAddr string

	// The listening address for secure GRPC. If the port in the address is empty or "0" (as in "127.0.0.1:" or "[::1]:0")
	// a port number is automatically chosen.
	SecureGrpcAddr string

	// The listening address for the monitoring port. If the port in the address is empty or "0" (as in "127.0.0.1:" or "[::1]:0")
	// a port number is automatically chosen.
	MonitoringAddr string

	EnableProfiling bool
	EnableCaching   bool
	WebhookEndpoint string
}

// NewDiscoveryService creates an Envoy discovery service on a given port
func NewDiscoveryService(ctl model.Controller, configCache model.ConfigStoreCache,
	environment *model.Environment, o DiscoveryServiceOptions) (*DiscoveryService, error) {
	out := &DiscoveryService{
		Environment: environment,
		sdsCache:    newDiscoveryCache("sds", o.EnableCaching),
	}

	container := restful.NewContainer()
	if o.EnableProfiling {
		container.ServeMux.HandleFunc("/debug/pprof/", pprof.Index)
		container.ServeMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		container.ServeMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		container.ServeMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		container.ServeMux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}
	out.Register(container)

	out.webhookEndpoint, out.webhookClient = util.NewWebHookClient(o.WebhookEndpoint)
	out.RestContainer = container

	// Flush cached discovery responses whenever services, service
	// instances, or routing configuration changes.
	serviceHandler := func(*model.Service, model.Event) { out.clearCache() }
	if err := ctl.AppendServiceHandler(serviceHandler); err != nil {
		return nil, err
	}
	instanceHandler := func(*model.ServiceInstance, model.Event) { out.clearCache() }
	if err := ctl.AppendInstanceHandler(instanceHandler); err != nil {
		return nil, err
	}

	// Flush cached discovery responses when detecting jwt public key change.
	model.JwtKeyResolver.PushFunc = out.ClearCache

	if configCache != nil {
		// TODO: changes should not trigger a full recompute of LDS/RDS/CDS/EDS
		// (especially mixerclient HTTP and quota)
		configHandler := func(model.Config, model.Event) {
			out.clearCache()
		}
		for _, descriptor := range model.IstioConfigTypes {
			configCache.RegisterEventHandler(descriptor.Type, configHandler)
		}
	}

	return out, nil
}

// Register adds routes a web service container. This is visible for testing purposes only.
func (ds *DiscoveryService) Register(container *restful.Container) {
	ws := &restful.WebService{}
	ws.Produces(restful.MIME_JSON)

	// List all known services (informational, not invoked by Envoy)
	ws.Route(ws.
		GET("/v1/registration").
		To(ds.ListAllEndpoints).
		Doc("Services in SDS"))

	ws.Route(ws.
		GET("/cache_stats").
		To(ds.GetCacheStats).
		Doc("Get discovery service cache stats").
		Writes(discoveryCacheStats{}))

	ws.Route(ws.
		POST("/cache_stats_delete").
		To(ds.ClearCacheStats).
		Doc("Clear discovery service cache stats"))

	container.Add(ws)
}

// GetCacheStats returns the statistics for cached discovery responses.
func (ds *DiscoveryService) GetCacheStats(_ *restful.Request, response *restful.Response) {
	stats := make(map[string]*discoveryCacheStatEntry)
	for k, v := range ds.sdsCache.stats() {
		stats[k] = v
	}
	if err := response.WriteEntity(discoveryCacheStats{stats}); err != nil {
		log.Warna(err)
	}
}

// ClearCacheStats clear the statistics for cached discovery responses.
func (ds *DiscoveryService) ClearCacheStats(_ *restful.Request, _ *restful.Response) {
	ds.sdsCache.resetStats()
}

// TODO: move debounce to v2/discovery, remove the spaghetti.

// ClearCache is wrapper for clearCache method, used when new controller gets
// instantiated dynamically
func (ds *DiscoveryService) ClearCache() {
	ds.clearCache()
}

// debouncePush is called on config or endpoint changes, to initiate a push.
func (ds *DiscoveryService) debouncePush(startDebounce time.Time) {
	clearCacheMutex.Lock()
	since := time.Since(lastClearCacheEvent)
	events := clearCacheEvents
	clearCacheMutex.Unlock()

	if since > 2*DebounceAfter ||
		time.Since(startDebounce) > DebounceMax {

		log.Infof("Push debounce stable %d: %v since last change, %v since last push, full=%v",
			events,
			since, time.Since(lastClearCache), ds.fullPush)

		ds.doPush()

	} else {
		time.AfterFunc(DebounceAfter, func() {
			ds.debouncePush(startDebounce)
		})
	}
}

// Start the actual push
func (ds *DiscoveryService) doPush() {
	// more config update events may happen while doPush is processing.
	// we don't want to lose updates.
	clearCacheMutex.Lock()

	clearCacheTimerSet = false
	lastClearCache = time.Now()
	full := ds.fullPush
	edsUpdates := BeforePush()

	// Update the config values, next ConfigUpdate and eds updates will use this
	ds.fullPush = false

	clearCacheMutex.Unlock()

	Push(full, edsUpdates)
}

// clearCache will clear all envoy caches. Called by service, instance and config handlers.
// This will impact the performance, since envoy will need to recalculate.
func (ds *DiscoveryService) clearCache() {
	ds.ConfigUpdate(true)
}

// ConfigUpdate implements ConfigUpdater interface, used to request pushes.
// It replaces the 'clear cache' from v1.
func (ds *DiscoveryService) ConfigUpdate(full bool) {
	clearCacheMutex.Lock()
	defer clearCacheMutex.Unlock()

	if full {
		ds.fullPush = true
	}
	clearCacheEvents++

	if DebounceAfter > 0 {
		lastClearCacheEvent = time.Now()

		if !clearCacheTimerSet {
			clearCacheTimerSet = true
			startDebounce := lastClearCacheEvent
			time.AfterFunc(DebounceAfter, func() {
				ds.debouncePush(startDebounce)
			})
		}
		return
	}

	// Old code, for safety
	// If last config change was > 1 second ago, push.
	if time.Since(lastClearCacheEvent) > 1*time.Second {
		log.Infof("Push %d: %v since last change, %v since last push",
			clearCacheEvents,
			time.Since(lastClearCacheEvent), time.Since(lastClearCache))
		lastClearCacheEvent = time.Now()

		ds.doPush()
		return
	}

	lastClearCacheEvent = time.Now()

	// If last config change was < 1 second ago, but last push is > clearCacheTime ago -
	// also push

	if time.Since(lastClearCache) > time.Duration(clearCacheTime)*time.Second {
		log.Infof("Timer push %d: %v since last change, %v since last push",
			clearCacheEvents, time.Since(lastClearCacheEvent), time.Since(lastClearCache))
		ds.doPush()
		return
	}

	// Last config change was < 1 second ago, and we're continuing to get changes.
	// Set a timer 1 second in the future, to evaluate again.
	// if a timer was already set, don't bother.
	if !clearCacheTimerSet {
		clearCacheTimerSet = true
		time.AfterFunc(1*time.Second, func() {
			clearCacheMutex.Lock()
			clearCacheTimerSet = false
			clearCacheMutex.Unlock()
			ds.clearCache() // re-evaluate after 1 second. If no activity - push will happen
		})
	}

}

// ListAllEndpoints responds with all Services and is not restricted to a single service-key
// Deprecated - may be used by debug tools, mapped to /v1/registration
func (ds *DiscoveryService) ListAllEndpoints(_ *restful.Request, response *restful.Response) {
	methodName := "ListAllEndpoints"
	incCalls(methodName)

	services := make([]*keyAndService, 0)

	svcs, err := ds.Services()
	if err != nil {
		// If client experiences an error, 503 error will tell envoy to keep its current
		// cache and try again later
		errorResponse(methodName, response, http.StatusServiceUnavailable, "EDS "+err.Error())
		return
	}

	for _, service := range svcs {
		if !service.External() {
			for _, port := range service.Ports {
				hosts := make([]*host, 0)
				instances, err := ds.Instances(service.Hostname, []string{port.Name}, nil)
				if err != nil {
					// If client experiences an error, 503 error will tell envoy to keep its current
					// cache and try again later
					errorResponse(methodName, response, http.StatusInternalServerError, "EDS "+err.Error())
					return
				}
				for _, instance := range instances {
					hosts = append(hosts, &host{
						Address: instance.Endpoint.Address,
						Port:    instance.Endpoint.Port,
					})
				}
				services = append(services, &keyAndService{
					Key:   service.Key(port, nil),
					Hosts: hosts,
				})
			}
		}
	}

	// Sort servicesArray.  This is not strictly necessary, but discovery_test.go will
	// be comparing against a golden example using test/util/diff.go which does a textual comparison
	sort.Slice(services, func(i, j int) bool { return services[i].Key < services[j].Key })

	if err := response.WriteEntity(services); err != nil {
		incErrors(methodName)
		log.Warna(err)
	} else {
		observeResources(methodName, uint32(len(services)))
	}
}

func (ds *DiscoveryService) parseDiscoveryRequest(request *restful.Request) (model.Proxy, error) {
	nodeInfo := request.PathParameter(ServiceNode)
	svcNode, err := model.ParseServiceNode(nodeInfo)
	if err != nil {
		return svcNode, multierror.Prefix(err, fmt.Sprintf("unexpected %s: ", ServiceNode))
	}
	return svcNode, nil
}

func (ds *DiscoveryService) invokeWebhook(path string, payload []byte, methodName string) ([]byte, error) {
	if ds.webhookClient == nil {
		return payload, nil
	}

	incWebhookCalls(methodName)
	resp, err := ds.webhookClient.Post(ds.webhookEndpoint+path, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		incWebhookErrors(methodName)
		return nil, err
	}

	defer resp.Body.Close() // nolint: errcheck

	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		incWebhookErrors(methodName)
	}

	return out, err
}

func incCalls(methodName string) {
	callCounter.With(prometheus.Labels{
		metricLabelMethod:  methodName,
		metricBuildVersion: buildVersion,
	}).Inc()
}

func incErrors(methodName string) {
	errorCounter.With(prometheus.Labels{
		metricLabelMethod:  methodName,
		metricBuildVersion: buildVersion,
	}).Inc()
}

func incWebhookCalls(methodName string) {
	webhookCallCounter.With(prometheus.Labels{
		metricLabelMethod:  methodName,
		metricBuildVersion: buildVersion,
	}).Inc()
}

func incWebhookErrors(methodName string) {
	webhookErrorCounter.With(prometheus.Labels{
		metricLabelMethod:  methodName,
		metricBuildVersion: buildVersion,
	}).Inc()
}

func observeResources(methodName string, count uint32) {
	resourceCounter.With(prometheus.Labels{
		metricLabelMethod:  methodName,
		metricBuildVersion: buildVersion,
	}).Observe(float64(count))
}

func errorResponse(methodName string, r *restful.Response, status int, msg string) {
	incErrors(methodName)
	log.Warn(msg)
	if err := r.WriteErrorString(status, msg); err != nil {
		log.Warna(err)
	}
}
