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

package bootstrap

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"code.cloudfoundry.org/copilot"
	"github.com/davecgh/go-spew/spew"
	durpb "github.com/golang/protobuf/ptypes/duration"
	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	multierror "github.com/hashicorp/go-multierror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	meshconfig "istio.io/api/mesh/v1alpha1"
	"istio.io/istio/pilot/cmd"
	configaggregate "istio.io/istio/pilot/pkg/config/aggregate"
	"istio.io/istio/pilot/pkg/config/clusterregistry"
	"istio.io/istio/pilot/pkg/config/kube/crd"
	"istio.io/istio/pilot/pkg/config/kube/ingress"
	"istio.io/istio/pilot/pkg/config/memory"
	configmonitor "istio.io/istio/pilot/pkg/config/monitor"
	"istio.io/istio/pilot/pkg/model"
	istio_networking "istio.io/istio/pilot/pkg/networking/core"
	"istio.io/istio/pilot/pkg/networking/plugin"
	"istio.io/istio/pilot/pkg/proxy/envoy"
	envoyv2 "istio.io/istio/pilot/pkg/proxy/envoy/v2"
	"istio.io/istio/pilot/pkg/serviceregistry"
	"istio.io/istio/pilot/pkg/serviceregistry/aggregate"
	"istio.io/istio/pilot/pkg/serviceregistry/cloudfoundry"
	"istio.io/istio/pilot/pkg/serviceregistry/consul"
	"istio.io/istio/pilot/pkg/serviceregistry/external"
	"istio.io/istio/pilot/pkg/serviceregistry/kube"
	srmemory "istio.io/istio/pilot/pkg/serviceregistry/memory"
	"istio.io/istio/pkg/ctrlz"
	"istio.io/istio/pkg/log"
	"istio.io/istio/pkg/version"
)

const (
	// ConfigMapKey should match the expected MeshConfig file name
	ConfigMapKey = "mesh"
	// CopilotTimeout when to cancel remote gRPC call to copilot
	CopilotTimeout = 5 * time.Second
)

var (
	// FilepathWalkInterval dictates how often the file system is walked for config
	FilepathWalkInterval = 100 * time.Millisecond

	// PilotCertDir is the default location for mTLS certificates used by pilot
	// Visible for tests - at runtime can be set by PILOT_CERT_DIR environment variable.
	PilotCertDir = "/etc/certs/"

	// DefaultPlugins is the default list of plugins to enable, when no plugin(s)
	// is specified through the command line
	DefaultPlugins = []string{
		plugin.Authn,
		plugin.Authz,
		plugin.Health,
		plugin.Mixer,
		plugin.Envoyfilter,
	}

	// Default is enabled - can be set to "0" to restore previous behavior, in case of problems.
	// Will be removed in 1.1 or 1.0.4 if we see no issues.
	directEDS = os.Getenv("PILOT_DIRECT_EDS") != "0"
)

// MeshArgs provide configuration options for the mesh. If ConfigFile is provided, an attempt will be made to
// load the mesh from the file. Otherwise, a default mesh will be used with optional overrides.
type MeshArgs struct {
	ConfigFile      string
	MixerAddress    string
	RdsRefreshDelay *durpb.Duration
}

// ConfigArgs provide configuration options for the configuration controller. If FileDir is set, that directory will
// be monitored for CRD yaml files and will update the controller as those files change (This is used for testing
// purposes). Otherwise, a CRD client is created based on the configuration.
type ConfigArgs struct {
	ClusterRegistriesConfigmap string
	ClusterRegistriesNamespace string
	KubeConfig                 string
	CFConfig                   string
	ControllerOptions          kube.ControllerOptions
	FileDir                    string
	DisableInstallCRDs         bool

	// Controller if specified, this controller overrides the other config settings.
	Controller model.ConfigStoreCache
}

// ConsulArgs provides configuration for the Consul service registry.
type ConsulArgs struct {
	Config    string
	ServerURL string
	Interval  time.Duration
}

// ServiceArgs provides the composite configuration for all service registries in the system.
type ServiceArgs struct {
	Registries []string
	Consul     ConsulArgs
}

// PilotArgs provides all of the configuration parameters for the Pilot discovery service.
type PilotArgs struct {
	DiscoveryOptions envoy.DiscoveryServiceOptions
	Namespace        string
	Mesh             MeshArgs
	Config           ConfigArgs
	Service          ServiceArgs
	MeshConfig       *meshconfig.MeshConfig
	CtrlZOptions     *ctrlz.Options
	Plugins          []string
}

// Server contains the runtime configuration for the Pilot discovery service.
type Server struct {
	HTTPListeningAddr       net.Addr
	GRPCListeningAddr       net.Addr
	SecureGRPCListeningAddr net.Addr
	MonitorListeningAddr    net.Addr

	// TODO(nmittler): Consider alternatives to exposing these directly
	EnvoyXdsServer    *envoyv2.DiscoveryServer
	ServiceController *aggregate.Controller

	mesh             *meshconfig.MeshConfig
	configController model.ConfigStoreCache
	mixerSAN         []string
	kubeClient       kubernetes.Interface
	startFuncs       []startFunc
	clusterStore     *clusterregistry.ClusterStore
	httpServer       *http.Server
	grpcServer       *grpc.Server
	secureGRPCServer *grpc.Server
	discoveryService *envoy.DiscoveryService
	istioConfigStore model.IstioConfigStore
	mux              *http.ServeMux
	kubeRegistry     *kube.Controller
}

func createInterface(kubeconfig string) (kubernetes.Interface, error) {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)

	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(restConfig)
}

// NewServer creates a new Server instance based on the provided arguments.
func NewServer(args PilotArgs) (*Server, error) {
	// If the namespace isn't set, try looking it up from the environment.
	if args.Namespace == "" {
		args.Namespace = os.Getenv("POD_NAMESPACE")
	}
	if args.Config.ClusterRegistriesNamespace == "" {
		if args.Namespace != "" {
			args.Config.ClusterRegistriesNamespace = args.Namespace
		} else {
			args.Config.ClusterRegistriesNamespace = "istio-system"
		}
	}

	s := &Server{}

	// Apply the arguments to the configuration.
	if err := s.initKubeClient(&args); err != nil {
		return nil, err
	}
	if err := s.initClusterRegistries(&args); err != nil {
		return nil, err
	}
	if err := s.initMesh(&args); err != nil {
		return nil, err
	}
	if err := s.initMixerSan(&args); err != nil {
		return nil, err
	}
	if err := s.initConfigController(&args); err != nil {
		return nil, err
	}
	if err := s.initServiceControllers(&args); err != nil {
		return nil, err
	}
	if err := s.initDiscoveryService(&args); err != nil {
		return nil, err
	}
	if err := s.initMonitor(&args); err != nil {
		return nil, err
	}
	if err := s.initMultiClusterController(&args); err != nil {
		return nil, err
	}

	if args.CtrlZOptions != nil {
		go ctrlz.Run(args.CtrlZOptions, nil)
	}

	return s, nil
}

// Start starts all components of the Pilot discovery service on the port specified in DiscoveryServiceOptions.
// If Port == 0, a port number is automatically chosen. This method returns the address on which the server is
// listening for incoming connections. Content serving is started by this method, but is executed asynchronously.
// Serving can be cancelled at any time by closing the provided stop channel.
func (s *Server) Start(stop <-chan struct{}) (net.Addr, error) {
	// Now start all of the components.
	for _, fn := range s.startFuncs {
		if err := fn(stop); err != nil {
			return nil, err
		}
	}

	return s.HTTPListeningAddr, nil
}

// startFunc defines a function that will be used to start one or more components of the Pilot discovery service.
type startFunc func(stop <-chan struct{}) error

// initMonitor initializes the configuration for the pilot monitoring server.
func (s *Server) initMonitor(args *PilotArgs) error {
	s.addStartFunc(func(stop <-chan struct{}) error {
		monitor, addr, err := startMonitor(args.DiscoveryOptions.MonitoringAddr, s.mux)
		if err != nil {
			return err
		}
		s.MonitorListeningAddr = addr

		go func() {
			<-stop
			err := monitor.Close()
			log.Debugf("Monitoring server terminated: %v", err)
		}()
		return nil
	})
	return nil
}

func (s *Server) initClusterRegistries(args *PilotArgs) (err error) {
	s.clusterStore = clusterregistry.NewClustersStore()

	if s.kubeClient == nil {
		log.Infof("skipping cluster registries, no kube-client created")
		return nil
	}

	// Drop from multicluster test cases if Mock Registry is used
	if checkForMock(args.Service.Registries) {
		return nil
	}
	if args.Config.ClusterRegistriesConfigmap != "" {
		if err = clusterregistry.ReadClusters(s.kubeClient,
			args.Config.ClusterRegistriesConfigmap,
			args.Config.ClusterRegistriesNamespace,
			s.clusterStore); err != nil {
			return err
		}
	}
	if s.clusterStore != nil {
		log.Infof("clusters configuration %s", spew.Sdump(s.clusterStore))
	}

	return err
}

// Check if Mock's registry exists in PilotArgs's Registries
func checkForMock(registries []string) bool {
	for _, r := range registries {
		if strings.ToLower(r) == "mock" {
			return true
		}
	}

	return false
}

// Check if Kubernetes registry exists in PilotArgs's Registries
func checkForKubernetes(registries []string) bool {
	for _, r := range registries {
		if strings.ToLower(r) == "kubernetes" {
			return true
		}
	}

	return false
}

// GetMeshConfig fetches the ProxyMesh configuration from Kubernetes ConfigMap.
func GetMeshConfig(kube kubernetes.Interface, namespace, name string) (*v1.ConfigMap, *meshconfig.MeshConfig, error) {

	if kube == nil {
		defaultMesh := model.DefaultMeshConfig()
		return nil, &defaultMesh, nil
	}

	config, err := kube.CoreV1().ConfigMaps(namespace).Get(name, meta_v1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("\"%s\" not found", name)) {
			defaultMesh := model.DefaultMeshConfig()
			return nil, &defaultMesh, nil
		}
		return nil, nil, err
	}

	// values in the data are strings, while proto might use a different data type.
	// therefore, we have to get a value by a key
	cfgYaml, exists := config.Data[ConfigMapKey]
	if !exists {
		return nil, nil, fmt.Errorf("missing configuration map key %q", ConfigMapKey)
	}

	mesh, err := model.ApplyMeshConfigDefaults(cfgYaml)
	if err != nil {
		return nil, nil, err
	}
	return config, mesh, nil
}

// initMesh creates the mesh in the pilotConfig from the input arguments.
func (s *Server) initMesh(args *PilotArgs) error {
	// If a config file was specified, use it.
	if args.MeshConfig != nil {
		s.mesh = args.MeshConfig
		return nil
	}
	var mesh *meshconfig.MeshConfig
	if args.Mesh.ConfigFile != "" {
		fileMesh, err := cmd.ReadMeshConfig(args.Mesh.ConfigFile)
		if err != nil {
			log.Warnf("failed to read mesh configuration, using default: %v", err)
		} else {
			mesh = fileMesh
		}
	}

	if mesh == nil {
		var err error
		// Config file either wasn't specified or failed to load - use a default mesh.
		if _, mesh, err = GetMeshConfig(s.kubeClient, kube.IstioNamespace, kube.IstioConfigMap); err != nil {
			log.Warnf("failed to read mesh configuration: %v", err)
			return err
		}

		// Allow some overrides for testing purposes.
		if args.Mesh.MixerAddress != "" {
			mesh.MixerCheckServer = args.Mesh.MixerAddress
			mesh.MixerReportServer = args.Mesh.MixerAddress
		}
		if args.Mesh.RdsRefreshDelay != nil {
			mesh.RdsRefreshDelay = args.Mesh.RdsRefreshDelay
		}
	}

	log.Infof("mesh configuration %s", spew.Sdump(mesh))
	log.Infof("version %s", version.Info.String())
	log.Infof("flags %s", spew.Sdump(args))

	s.mesh = mesh
	return nil
}

// initMixerSan configures the mixerSAN configuration item. The mesh must already have been configured.
func (s *Server) initMixerSan(args *PilotArgs) error {
	if s.mesh == nil {
		return fmt.Errorf("the mesh has not been configured before configuring mixer san")
	}
	if s.mesh.DefaultConfig.ControlPlaneAuthPolicy == meshconfig.AuthenticationPolicy_MUTUAL_TLS {
		s.mixerSAN = envoy.GetMixerSAN(args.Config.ControllerOptions.DomainSuffix, args.Namespace)
	}
	return nil
}

func (s *Server) getKubeCfgFile(args *PilotArgs) (kubeCfgFile string) {
	if kubeCfgFile == "" {
		kubeCfgFile = args.Config.KubeConfig
	}
	return
}

// initKubeClient creates the k8s client if running in an k8s environment.
func (s *Server) initKubeClient(args *PilotArgs) error {
	var needToCreateClient bool
	for _, r := range args.Service.Registries {
		if serviceregistry.ServiceRegistry(r) == serviceregistry.KubernetesRegistry {
			needToCreateClient = true
			break
		}
	}

	if needToCreateClient && args.Config.FileDir == "" {
		var client kubernetes.Interface
		var kuberr error

		kubeCfgFile := s.getKubeCfgFile(args)
		client, kuberr = createInterface(kubeCfgFile)

		if kuberr != nil {
			return multierror.Prefix(kuberr, "failed to connect to Kubernetes API.")
		}
		s.kubeClient = client
	}
	return nil
}

type mockController struct{}

func (c *mockController) AppendServiceHandler(f func(*model.Service, model.Event)) error {
	return nil
}

func (c *mockController) AppendInstanceHandler(f func(*model.ServiceInstance, model.Event)) error {
	return nil
}

func (c *mockController) Run(<-chan struct{}) {}

// initConfigController creates the config controller in the pilotConfig.
func (s *Server) initConfigController(args *PilotArgs) error {
	if args.Config.Controller != nil {
		s.configController = args.Config.Controller
	} else if args.Config.FileDir != "" {
		store := memory.Make(model.IstioConfigTypes)
		configController := memory.NewController(store)

		err := s.makeFileMonitor(args, configController)
		if err != nil {
			return err
		}

		if args.Config.CFConfig != "" {
			err = s.makeCopilotMonitor(args, configController)
			if err != nil {
				return err
			}
		}

		s.configController = configController
	} else {
		controller, err := s.makeKubeConfigController(args)
		if err != nil {
			return err
		}

		s.configController = controller
	}

	// Defer starting the controller until after the service is created.
	s.addStartFunc(func(stop <-chan struct{}) error {
		go s.configController.Run(stop)
		return nil
	})

	// If running in ingress mode (requires k8s), wrap the config controller.
	if hasKubeRegistry(args) && s.mesh.IngressControllerMode != meshconfig.MeshConfig_OFF {
		// Wrap the config controller with a cache.
		configController, err := configaggregate.MakeCache([]model.ConfigStoreCache{
			s.configController,
			ingress.NewController(s.kubeClient, s.mesh, args.Config.ControllerOptions),
		})
		if err != nil {
			return err
		}

		// Update the config controller
		s.configController = configController

		if ingressSyncer, errSyncer := ingress.NewStatusSyncer(s.mesh, s.kubeClient,
			args.Namespace, args.Config.ControllerOptions); errSyncer != nil {
			log.Warnf("Disabled ingress status syncer due to %v", errSyncer)
		} else {
			s.addStartFunc(func(stop <-chan struct{}) error {
				go ingressSyncer.Run(stop)
				return nil
			})
		}
	}

	// Create the config store.
	s.istioConfigStore = model.MakeIstioStore(s.configController)

	return nil
}

func (s *Server) makeKubeConfigController(args *PilotArgs) (model.ConfigStoreCache, error) {
	kubeCfgFile := s.getKubeCfgFile(args)
	configClient, err := crd.NewClient(kubeCfgFile, "", model.IstioConfigTypes, args.Config.ControllerOptions.DomainSuffix)
	if err != nil {
		return nil, multierror.Prefix(err, "failed to open a config client.")
	}

	if !args.Config.DisableInstallCRDs {
		if err = configClient.RegisterResources(); err != nil {
			return nil, multierror.Prefix(err, "failed to register custom resources.")
		}
	}

	return crd.NewController(configClient, args.Config.ControllerOptions), nil
}

func (s *Server) makeFileMonitor(args *PilotArgs, configController model.ConfigStore) error {
	fileSnapshot := configmonitor.NewFileSnapshot(args.Config.FileDir, model.IstioConfigTypes)
	fileMonitor := configmonitor.NewMonitor(configController, FilepathWalkInterval, fileSnapshot.ReadConfigFiles)

	// Defer starting the file monitor until after the service is created.
	s.addStartFunc(func(stop <-chan struct{}) error {
		fileMonitor.Start(stop)
		return nil
	})

	return nil
}

func (s *Server) makeCopilotMonitor(args *PilotArgs, configController model.ConfigStore) error {
	cfConfig, err := cloudfoundry.LoadConfig(args.Config.CFConfig)
	if err != nil {
		return multierror.Prefix(err, "loading cloud foundry config")
	}
	tlsConfig, err := cfConfig.ClientTLSConfig()
	if err != nil {
		return multierror.Prefix(err, "creating cloud foundry client tls config")
	}
	client, err := copilot.NewIstioClient(cfConfig.Copilot.Address, tlsConfig)
	if err != nil {
		return multierror.Prefix(err, "creating cloud foundry client")
	}

	copilotSnapshot := configmonitor.NewCopilotSnapshot(configController, client, CopilotTimeout)
	copilotMonitor := configmonitor.NewMonitor(configController, 1*time.Second, copilotSnapshot.ReadConfigFiles)

	s.addStartFunc(func(stop <-chan struct{}) error {
		copilotMonitor.Start(stop)
		return nil
	})

	return nil
}

// createK8sServiceControllers creates all the k8s service controllers under this pilot
func (s *Server) createK8sServiceControllers(serviceControllers *aggregate.Controller, args *PilotArgs) (err error) {
	clusterID := string(serviceregistry.KubernetesRegistry)
	log.Infof("Primary Cluster name: %s", clusterID)
	args.Config.ControllerOptions.ClusterID = clusterID
	kubectl := kube.NewController(s.kubeClient, args.Config.ControllerOptions)
	s.kubeRegistry = kubectl
	serviceControllers.AddRegistry(
		aggregate.Registry{
			Name:             serviceregistry.KubernetesRegistry,
			ClusterID:        clusterID,
			ServiceDiscovery: kubectl,
			ServiceAccounts:  kubectl,
			Controller:       kubectl,
		})

	return
}

// initMultiClusterController initializes multi cluster controller
// currently implemented only for kubernetes registries
func (s *Server) initMultiClusterController(args *PilotArgs) (err error) {
	if checkForKubernetes(args.Service.Registries) {
		// Start secret controller which watches for runtime secret Object changes and adds secrets dynamically
		err = clusterregistry.StartSecretController(s.kubeClient,
			s.clusterStore,
			s.ServiceController,
			s.discoveryService,
			s.EnvoyXdsServer,
			args.Config.ClusterRegistriesNamespace,
			args.Config.ControllerOptions.ResyncPeriod,
			args.Config.ControllerOptions.WatchedNamespace,
			args.Config.ControllerOptions.DomainSuffix)
	}
	return
}

func hasKubeRegistry(args *PilotArgs) bool {
	for _, r := range args.Service.Registries {
		if serviceregistry.ServiceRegistry(r) == serviceregistry.KubernetesRegistry {
			return true
		}
	}
	return false
}

// initServiceControllers creates and initializes the service controllers
func (s *Server) initServiceControllers(args *PilotArgs) error {
	serviceControllers := aggregate.NewController()
	registered := make(map[serviceregistry.ServiceRegistry]bool)
	for _, r := range args.Service.Registries {
		serviceRegistry := serviceregistry.ServiceRegistry(r)
		if _, exists := registered[serviceRegistry]; exists {
			log.Warnf("%s registry specified multiple times.", r)
			continue
		}
		registered[serviceRegistry] = true
		log.Infof("Adding %s registry adapter", serviceRegistry)
		switch serviceRegistry {
		case serviceregistry.ConfigRegistry:
			s.initConfigRegistry(serviceControllers)
		case serviceregistry.MockRegistry:
			s.initMemoryRegistry(serviceControllers)
		case serviceregistry.KubernetesRegistry:
			if err := s.createK8sServiceControllers(serviceControllers, args); err != nil {
				return err
			}
		case serviceregistry.ConsulRegistry:
			log.Infof("Consul url: %v", args.Service.Consul.ServerURL)
			conctl, conerr := consul.NewController(
				args.Service.Consul.ServerURL, args.Service.Consul.Interval)
			if conerr != nil {
				return fmt.Errorf("failed to create Consul controller: %v", conerr)
			}
			serviceControllers.AddRegistry(
				aggregate.Registry{
					Name:             serviceregistry.ServiceRegistry(r),
					ServiceDiscovery: conctl,
					ServiceAccounts:  conctl,
					Controller:       conctl,
				})

		case serviceregistry.CloudFoundryRegistry:
			cfConfig, err := cloudfoundry.LoadConfig(args.Config.CFConfig)
			if err != nil {
				return multierror.Prefix(err, "loading cloud foundry config")
			}
			tlsConfig, err := cfConfig.ClientTLSConfig()
			if err != nil {
				return multierror.Prefix(err, "creating cloud foundry client tls config")
			}
			client, err := copilot.NewIstioClient(cfConfig.Copilot.Address, tlsConfig)
			if err != nil {
				return multierror.Prefix(err, "creating cloud foundry client")
			}
			serviceControllers.AddRegistry(aggregate.Registry{
				Name: serviceregistry.ServiceRegistry(r),
				Controller: &cloudfoundry.Controller{
					Ticker: cloudfoundry.NewTicker(cfConfig.Copilot.PollInterval),
					Client: client,
				},
				ServiceDiscovery: &cloudfoundry.ServiceDiscovery{
					Client:      client,
					ServicePort: cfConfig.ServicePort,
				},
				ServiceAccounts: cloudfoundry.NewServiceAccounts(),
			})

		default:
			return multierror.Prefix(nil, "Service registry "+r+" is not supported.")
		}
	}
	serviceEntryStore := external.NewServiceDiscovery(s.configController, s.istioConfigStore)

	// add service entry registry to aggregator by default
	serviceEntryRegistry := aggregate.Registry{
		Name:             "ServiceEntries",
		Controller:       serviceEntryStore,
		ServiceDiscovery: serviceEntryStore,
		ServiceAccounts:  serviceEntryStore,
	}
	serviceControllers.AddRegistry(serviceEntryRegistry)

	s.ServiceController = serviceControllers

	// Defer running of the service controllers.
	s.addStartFunc(func(stop <-chan struct{}) error {
		go s.ServiceController.Run(stop)
		return nil
	})

	return nil
}

func (s *Server) initMemoryRegistry(serviceControllers *aggregate.Controller) {
	// MemServiceDiscovery implementation
	discovery1 := srmemory.NewDiscovery(
		map[model.Hostname]*model.Service{ // srmemory.HelloService.Hostname: srmemory.HelloService,
		}, 2)

	discovery2 := srmemory.NewDiscovery(
		map[model.Hostname]*model.Service{ // srmemory.WorldService.Hostname: srmemory.WorldService,
		}, 2)

	registry1 := aggregate.Registry{
		Name:             serviceregistry.ServiceRegistry("mockAdapter1"),
		ClusterID:        "mockAdapter1",
		ServiceDiscovery: discovery1,
		ServiceAccounts:  discovery1,
		Controller:       &mockController{},
	}

	registry2 := aggregate.Registry{
		Name:             serviceregistry.ServiceRegistry("mockAdapter2"),
		ClusterID:        "mockAdapter2",
		ServiceDiscovery: discovery2,
		ServiceAccounts:  discovery2,
		Controller:       &mockController{},
	}
	serviceControllers.AddRegistry(registry1)
	serviceControllers.AddRegistry(registry2)
}

func (s *Server) initConfigRegistry(serviceControllers *aggregate.Controller) {
	serviceEntryStore := external.NewServiceDiscovery(s.configController, s.istioConfigStore)
	serviceControllers.AddRegistry(aggregate.Registry{
		Name:             serviceregistry.ConfigRegistry,
		ServiceDiscovery: serviceEntryStore,
		ServiceAccounts:  serviceEntryStore,
		Controller:       serviceEntryStore,
	})
}

func (s *Server) initDiscoveryService(args *PilotArgs) error {
	environment := &model.Environment{
		Mesh:             s.mesh,
		IstioConfigStore: s.istioConfigStore,
		ServiceDiscovery: s.ServiceController,
		ServiceAccounts:  s.ServiceController,
		MixerSAN:         s.mixerSAN,
	}

	// Set up discovery service
	discovery, err := envoy.NewDiscoveryService(
		s.ServiceController,
		s.configController,
		environment,
		args.DiscoveryOptions,
	)
	if err != nil {
		return fmt.Errorf("failed to create discovery service: %v", err)
	}
	s.discoveryService = discovery

	s.mux = s.discoveryService.RestContainer.ServeMux

	// For now we create the gRPC server sourcing data from Pilot's older data model.
	s.initGrpcServer()

	s.EnvoyXdsServer = envoyv2.NewDiscoveryServer(environment, istio_networking.NewConfigGenerator(args.Plugins))
	s.EnvoyXdsServer.ConfigUpdater = s.discoveryService
	// TODO: decouple v2 from the cache invalidation, use direct listeners.
	envoy.Push = s.EnvoyXdsServer.Push
	envoy.BeforePush = s.EnvoyXdsServer.BeforePush

	s.EnvoyXdsServer.Register(s.grpcServer)

	if s.kubeRegistry != nil {
		// kubeRegistry may use the environment for push status reporting.
		// TODO: maybe all registries should have his as an optional field ?
		s.kubeRegistry.Env = environment
		s.kubeRegistry.ConfigUpdater = discovery
		if directEDS {
			s.kubeRegistry.EDSUpdater = s.EnvoyXdsServer
		}
	}

	s.EnvoyXdsServer.InitDebug(s.mux, s.ServiceController, discovery)

	s.EnvoyXdsServer.ConfigController = s.configController

	s.httpServer = &http.Server{
		Addr:    args.DiscoveryOptions.HTTPAddr,
		Handler: discovery.RestContainer}

	listener, err := net.Listen("tcp", args.DiscoveryOptions.HTTPAddr)
	if err != nil {
		return err
	}
	s.HTTPListeningAddr = listener.Addr()

	grpcListener, err := net.Listen("tcp", args.DiscoveryOptions.GrpcAddr)
	if err != nil {
		return err
	}
	s.GRPCListeningAddr = grpcListener.Addr()

	// TODO: only if TLS certs, go routine to check for late certs
	secureGrpcListener, err := net.Listen("tcp", args.DiscoveryOptions.SecureGrpcAddr)
	if err != nil {
		return err
	}
	s.SecureGRPCListeningAddr = secureGrpcListener.Addr()

	s.addStartFunc(func(stop <-chan struct{}) error {
		log.Infof("Discovery service started at http=%s grpc=%s", listener.Addr().String(), grpcListener.Addr().String())

		go func() {
			if err = s.httpServer.Serve(listener); err != nil {
				log.Warna(err)
			}
		}()
		go func() {
			if err = s.grpcServer.Serve(grpcListener); err != nil {
				log.Warna(err)
			}
		}()
		if len(args.DiscoveryOptions.SecureGrpcAddr) > 0 {
			go s.secureGrpcStart(secureGrpcListener)
		}

		go func() {
			<-stop
			model.JwtKeyResolver.Close()

			err = s.httpServer.Close()
			if err != nil {
				log.Warna(err)
			}
			s.grpcServer.Stop()
			if s.secureGRPCServer != nil {
				s.secureGRPCServer.Stop()
			}
		}()

		return err
	})

	return nil
}

func (s *Server) initGrpcServer() {
	grpcOptions := s.grpcServerOptions()
	s.grpcServer = grpc.NewServer(grpcOptions...)
}

// The secure grpc will start when the credentials are found.
func (s *Server) secureGrpcStart(listener net.Listener) {
	certDir := os.Getenv("PILOT_CERT_DIR")
	if certDir == "" {
		certDir = PilotCertDir // /etc/certs
	}
	if !strings.HasSuffix(certDir, "/") {
		certDir = certDir + "/"
	}

	for i := 0; i < 30; i++ {
		opts := s.grpcServerOptions()

		// This is used for the grpc h2 implementation. It doesn't appear to be needed in
		// the case of golang h2 stack.
		creds, err := credentials.NewServerTLSFromFile(certDir+model.CertChainFilename,
			certDir+model.KeyFilename)
		// certs not ready yet.
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		// TODO: parse the file to determine expiration date. Restart listener before expiration
		cert, err := tls.LoadX509KeyPair(certDir+model.CertChainFilename,
			certDir+model.KeyFilename)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		caCertFile := certDir + model.RootCertFilename
		caCert, err := ioutil.ReadFile(caCertFile)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		opts = append(opts, grpc.Creds(creds))
		s.secureGRPCServer = grpc.NewServer(opts...)

		s.EnvoyXdsServer.Register(s.secureGRPCServer)

		log.Infof("Starting GRPC secure on %v with certs in %s", listener.Addr(), certDir)

		s := &http.Server{
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
				VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
					// For now accept any certs - pilot is not authenticating the caller, TLS used for
					// privacy
					return nil
				},
				NextProtos: []string{"h2", "http/1.1"},
				//ClientAuth: tls.NoClientCert,
				//ClientAuth: tls.RequestClientCert,
				ClientAuth: tls.RequireAndVerifyClientCert,
				ClientCAs:  caCertPool,
			},
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.ProtoMajor == 2 && strings.HasPrefix(
					r.Header.Get("Content-Type"), "application/grpc") {
					s.secureGRPCServer.ServeHTTP(w, r)
				} else {
					s.mux.ServeHTTP(w, r)
				}
			}),
		}

		// This seems the only way to call setupHTTP2 - it may also be possible to set NextProto
		// on a listener
		_ = s.ServeTLS(listener, certDir+model.CertChainFilename, certDir+model.KeyFilename)

		// The other way to set TLS - but you can't add http handlers, and the h2 stack is
		// different.
		//if err := s.secureGRPCServer.Serve(listener); err != nil {
		//	log.Warna(err)
		//}
	}

	log.Errorf("Failed to find certificates for GRPC secure in %s", certDir)

	// Exit - mesh is in MTLS mode, but certificates are missing or bad.
	// k8s may allocate to a different machine.
	if s.mesh.DefaultConfig.ControlPlaneAuthPolicy == meshconfig.AuthenticationPolicy_MUTUAL_TLS {
		os.Exit(403)
	}
}

func (s *Server) grpcServerOptions() []grpc.ServerOption {
	interceptors := []grpc.UnaryServerInterceptor{
		// setup server prometheus monitoring (as final interceptor in chain)
		prometheus.UnaryServerInterceptor,
	}

	prometheus.EnableHandlingTimeHistogram()

	// Temp setting, default should be enough for most supported environments. Can be used for testing
	// envoy with lower values.
	var maxStreams int
	maxStreamsEnv := os.Getenv("ISTIO_GPRC_MAXSTREAMS")
	if len(maxStreamsEnv) > 0 {
		maxStreams, _ = strconv.Atoi(maxStreamsEnv)
	}
	if maxStreams == 0 {
		maxStreams = 100000
	}

	grpcOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(middleware.ChainUnaryServer(interceptors...)),
		grpc.MaxConcurrentStreams(uint32(maxStreams)),
	}

	// get the grpc server wired up
	grpc.EnableTracing = true

	return grpcOptions
}

func (s *Server) addStartFunc(fn startFunc) {
	s.startFuncs = append(s.startFuncs, fn)
}
