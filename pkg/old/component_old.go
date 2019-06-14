package component

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
	"time"

	"github.com/ostromart/istio-installer/pkg/translate"

	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/ghodss/yaml"
	"github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
	"github.com/ostromart/istio-installer/pkg/helm"
	"github.com/ostromart/istio-installer/pkg/kube"
	"github.com/ostromart/istio-installer/pkg/kubectlcmd"
	"github.com/ostromart/istio-installer/pkg/manifest"
	"github.com/ostromart/istio-installer/pkg/util"
	"istio.io/istio/pkg/log"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/helm/pkg/chartutil"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ComponentState is the current state of a component.
type ComponentState string

const (
	ComponentStateNonExistent ComponentState = "NonExistent"
	ComponentStateUpdating    ComponentState = "Updating"
	ComponentStateHealthy     ComponentState = "Healthy"
	ComponentStateError       ComponentState = "Error"

	istioAPIVersion = "install.istio.io/v1alpha1"

	// applyLoopTime is the interval between kubectl applying the latest manifest for a component.
	applyLoopTime = time.Minute
	// cRDPollInterval is how often the state of CRDs is polled when waiting for their creation.
	cRDPollInterval = 500 * time.Millisecond
	// cRDPollTimeout is the maximum wait time for all CRDs to be created.
	cRDPollTimeout = 60 * time.Second

	// helmValuesFile is the default name of the values file.
	helmValuesFile = "values.yaml"
)

type Component interface {
	RenderManifest() (string, error)
}

type PilotComponent struct {
}

func (c *PilotComponent) RenderManifest() (string, error) {
	return "", nil
}

func NewPilotComponent(installSpec *v1alpha1.InstallSpec) *PilotComponent {
	return nil
}

type ProxyComponent struct {
}

func (c *ProxyComponent) RenderManifest() (string, error) {
	return "", nil
}

func NewProxyComponent(installSpec *v1alpha1.InstallSpec) *ProxyComponent {
	return nil
}

type SidecarComponent struct {
}

func (c *SidecarComponent) RenderManifest() (string, error) {
	return "", nil
}

func NewSidecarComponent(installSpec *v1alpha1.InstallSpec) *SidecarComponent {
	return nil
}

type CitadelComponent struct {
}

func (c *CitadelComponent) RenderManifest() (string, error) {
	return "", nil
}

func NewCitadelComponent(installSpec *v1alpha1.InstallSpec) *CitadelComponent {
	return nil
}

type CertManagerComponent struct {
}

func (c *CertManagerComponent) RenderManifest() (string, error) {
	return "", nil
}

func NewCertManagerComponent(installSpec *v1alpha1.InstallSpec) *CertManagerComponent {
	return nil
}

type NodeAgentComponent struct {
}

func (c *NodeAgentComponent) RenderManifest() (string, error) {
	return "", nil
}

func NewNodeAgentComponent(installSpec *v1alpha1.InstallSpec) *NodeAgentComponent {
	return nil
}

// ComponentDeploymentConfig is a configuration for a deployment component.
type ComponentDeploymentConfig struct {
	// ComponentName is the name of the component.
	ComponentName string
	// Namespace is the namespace for the component.
	Namespace string
	// Version is the version for the component e.g. 1.0.2
	Version string
	// Dependencies is a list of dependencies. Each dependency must reach the listed Version before the component
	// manifest can be applied.
	Dependencies []*Dependency
	// HelmChartBaseDirPath is the path to the root of the helm chart templates.
	HelmChartBaseDirPath string
	// RelativeChartPath is the relative path of the component chart from HelmChartBaseDirPath.
	RelativeChartPath string
	// Client is the kubernetes client.
	Client client.Client
	// Config is the config for kubernetes client.
	Config *rest.Config
	// Clientset is the kubernetes Clientset.
	Clientset *kubernetes.Clientset
}

// Dependency defines a dependency for the component.
type Dependency struct {
	DeploymentComponent *DeploymentComponent
	Version             string
}

// OwnerObject is an owner object used for GC.
type OwnerObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

// NewComponentDeployment creates a new DeploymentComponent with the given config and returns a pointer to it.
func NewComponentDeployment(cfg *ComponentDeploymentConfig) *DeploymentComponent {
	return &DeploymentComponent{
		componentName:        cfg.ComponentName,
		componentHash:        util.RandomString(20),
		namespace:            cfg.Namespace,
		version:              cfg.Version,
		state:                ComponentStateNonExistent,
		componentObject:      &OwnerObject{},
		dependencies:         cfg.Dependencies,
		helmChartBaseDirPath: cfg.HelmChartBaseDirPath,
		relativeChartPath:    cfg.RelativeChartPath,
		client:               cfg.Client,
		config:               cfg.Config,
		clientset:            cfg.Clientset,
		notifyChs:            make(map[string]map[chan<- struct{}]bool),
		kubectl:              kubectlcmd.New(),
	}
}

type DeploymentComponent struct {
	componentName string
	componentHash string
	namespace     string
	version       string

	state           ComponentState
	currentManifest string
	componentObject *OwnerObject

	dependencies         []*Dependency
	helmChartBaseDirPath string
	relativeChartPath    string
	helmValues           <-chan string
	helmTemplateRenderer *helm.TemplateRenderer

	client    client.Client
	kubectl   *kubectlcmd.Client
	config    *rest.Config
	clientset *kubernetes.Clientset

	updateDepsCh chan struct{}
	notifyChs    map[string]map[chan<- struct{}]bool
	notifyChsMu  sync.RWMutex
}

// Name is the name of the component.
func (d *DeploymentComponent) Name() string {
	return d.componentName
}

// State is the current state of the component.
func (d *DeploymentComponent) State() ComponentState {
	return d.state
}

// Manifest is the rendered YAML manifest of the component.
func (d *DeploymentComponent) Manifest() string {
	return d.currentManifest
}

// Version is the current Version of the component.
func (d *DeploymentComponent) Version() string {
	return d.version
}

// Dependencies returns the current dependencies of the component.
func (d *DeploymentComponent) Dependencies() []*Dependency {
	return d.dependencies
}

func (d *DeploymentComponent) setVersion(version string) {
	d.notifyChsMu.RLock()
	defer d.notifyChsMu.RUnlock()

	for ch := range d.notifyChs[version] {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

func (d *DeploymentComponent) SetDeps(deps []*Dependency) {
	d.dependencies = deps
	// Make this non blocking since if we aren't waiting on this signal, we don't care about dependencies.
	select {
	case d.updateDepsCh <- struct{}{}:
	default:
	}
}

func (d *DeploymentComponent) waitDeps() {
	if len(d.dependencies) > 0 {
		log.Infof("Component %s waiting for dependencies:", d.Name())
	}
	for _, dep := range d.dependencies {
		log.Infof("  %s", dep.DeploymentComponent.Name())
		ch := dep.DeploymentComponent.AddNotifyCond(dep.Version)
		<-ch
		log.Infof("  Dependency %s done.", dep.DeploymentComponent.Name())
	}
}

// AddNotifyCond returns channel that emits a signal when the component Version equals Version.
func (d *DeploymentComponent) AddNotifyCond(version string) <-chan struct{} {
	d.notifyChsMu.Lock()
	defer d.notifyChsMu.Unlock()
	if d.notifyChs[version] == nil {
		d.notifyChs[version] = make(map[chan<- struct{}]bool)
	}
	ch := make(chan struct{}, 1)
	d.notifyChs[version][ch] = true
	if d.version == version {
		ch <- struct{}{}
	}
	return ch
}

// RunApplyLoop immediately applies the rendered manifest and thereafter listens for changes in the manifest
// templates and the values ConfigMap and immediately reapplies a new manifest whenever these change.
// It also periodically reapplies the current manifest to reconcile any changes made to the API server outside
// of the operator config.
func (d *DeploymentComponent) RunApplyLoop(ctx context.Context) error {
	d.waitDeps()

	_, err := kube.GetValues(d.clientset, d.namespace, d.Name())
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	// TODO: put in a real reconciler here.
	ticker := time.NewTicker(applyLoopTime)

	d.helmTemplateRenderer = helm.NewHelmTemplateRenderer(d.helmChartBaseDirPath, d.componentName, d.namespace, d.config, d.clientset)
	if err := d.helmTemplateRenderer.Run(); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			d.helmTemplateRenderer.RenderNow()
		case mstr := <-d.helmTemplateRenderer.ManifestCh:
			if err := d.applyManifest(ctx, mstr); err != nil {
				log.Error(err.Error())
			}
		case <-d.updateDepsCh:
			d.waitDeps()
			d.helmTemplateRenderer.RenderNow()
		}
	}

	d.helmTemplateRenderer.RenderNow()

	return nil
}

// ApplyOnce renders and applies the manifest and returns.
func (d *DeploymentComponent) ApplyOnce(ctx context.Context, is *v1alpha1.InstallSpec) error {
	mstr, err := d.Render(is)
	if err != nil {
		return err
	}
	return d.applyManifest(ctx, mstr)
}

// RenderToDir returns the rendered manifest.
func (d *DeploymentComponent) Render(is *v1alpha1.InstallSpec) (string, error) {
	d.waitDeps()
	globalValues, err := getHelmValues(d.helmChartBaseDirPath)
	if err != nil {
		return "", err
	}

	subchartDirPath := filepath.Join(d.helmChartBaseDirPath, d.relativeChartPath, d.componentName)
	chart, err := chartutil.Load(subchartDirPath)
	if err != nil {
		return "", err
	}

	values := globalValues
	if filepath.Clean(d.helmChartBaseDirPath) != filepath.Clean(subchartDirPath) {
		var err error
		values, err = getHelmValues(subchartDirPath)
		if err != nil {
			return "", err
		}
		values, err = patchValues(bytes.Join([][]byte{globalValues, values}, nil), is)
	}
	//	log.Infof("values: \n%s\n", values)

	log.Infof("Rendering %s", subchartDirPath)
	baseYAML, err := helm.renderChart(d.namespace, string(values), chart)
	if err != nil {
		return "", err
	}

	return baseYAML, err
}

func patchYAMLManifest(baseYAML string, resourceOverride []*unstructured.Unstructured) (string, error) {
	baseObjs, err := manifest.ParseObjectsFromYAMLManifest(context.TODO(), baseYAML)
	if err != nil {
		return "", err
	}
	overrideObjs, err := manifest.ObjectsFromUnstructuredSlice(resourceOverride)
	if err != nil {
		return "", err
	}

	bom, oom := baseObjs.ToMap(), overrideObjs.ToMap()
	var ret strings.Builder

	// Try to apply the defined overlays.
	for k, oo := range oom {
		bo := bom[k]
		if bo == nil {
			// TODO: error log overlays with no matches in any component.
			continue
		}
		boj, err := bo.JSON()
		if err != nil {
			log.Errorf("JSON error (%s) for base manifest object: \n%q", err, bo)
			continue
		}
		ooj, err := oo.JSON()
		if err != nil {
			log.Errorf("JSON error (%s) for override object: \n%q", err, oo)
			continue
		}
		merged, err := jsonpatch.MergePatch(boj, ooj)
		if err != nil {
			log.Errorf("JSON merge error (%s) for base object: \n%s\n override object: \n%s", err, boj, ooj)
			continue
		}
		my, err := yaml.JSONToYAML(merged)
		if err != nil {
			log.Errorf("JSONToYAML error (%s) for merged object: \n%s", err, merged)
			continue
		}
		//log.Infof("Base object: \n%s\nAfter overlay:\n%s", bo.YAMLDebugString(), merged)
		ret.Write(my)
		ret.WriteString("\n---\n")
	}
	// renderChart the remaining objects with no overlays.
	for k, oo := range bom {
		if oom[k] != nil {
			// Skip objects that have overlays.
			continue
		}
		oy, err := oo.YAML()
		if err != nil {
			log.Errorf("Object to YAML error (%s) for base object: \n%q", err, oo)
			continue
		}
		ret.Write(oy)
		ret.WriteString("\n---\n")
	}
	return ret.String(), nil
}

func patchValues(baseYAML []byte, is *v1alpha1.InstallSpec) ([]byte, error) {
	baseValuesJSON, err := yaml.YAMLToJSON([]byte(baseYAML))
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal baseYAML to JSON: %s", err)
	}

	crValues, err := translate.ProtoToValues(is)
	if err != nil {
		return nil, err
	}
	overrideValuesJSON, err := yaml.YAMLToJSON([]byte(crValues))
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal baseYAML to JSON: %s", err)
	}

	merged, err := jsonpatch.MergePatch(baseValuesJSON, overrideValuesJSON)
	if err != nil {
		return nil, fmt.Errorf("merge patch failed: %s", err)
	}

	return yaml.JSONToYAML(merged)
}

func getHelmValues(helmChartDirPath string) ([]byte, error) {
	b, err := ioutil.ReadFile(filepath.Join(helmChartDirPath, helmValuesFile))
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (d *DeploymentComponent) applyManifest(ctx context.Context, manifestStr string) error {
	objects, err := manifest.ParseObjectsFromYAMLManifest(ctx, manifestStr)
	if err != nil {
		return err
	}
	err = d.injectOwnerRef(ctx, objects)
	if err != nil {
		return err
	}

	for _, o := range objects.Items {
		o.AddLabels(map[string]string{"component": d.componentName})
	}
	objects.Sort(defaultObjectOrder())

	extraArgs := []string{"--force", "--prune", "--selector", fmt.Sprintf("component=%s", d.componentName)}

	crdObjects := cRDKindObjects(objects)

	mcrd, err := crdObjects.JSONManifest()
	if err != nil {
		return err
	}
	if err := d.kubectl.Apply(ctx, d.namespace, mcrd, extraArgs...); err != nil {
		return err
	}
	// Not all Istio components are robust to not yet created CRDs.
	if err := d.waitForCRDs(objects); err != nil {
		return err
	}

	m, err := objects.JSONManifest()
	if err != nil {
		return err
	}
	if err := d.kubectl.Apply(ctx, d.namespace, m, extraArgs...); err != nil {
		return err
	}

	/*	if err := d.ensureWatches(ctx, name, objects); err != nil {
			klog.FromContext(ctx).Error().Err(err).Message("watching deployed object types")
			panic(fmt.Errorf("error watching deployed object types: %v", err))
		}
	*/
	return nil
}

func (d *DeploymentComponent) injectOwnerRef(ctx context.Context, manifestObjects *manifest.Objects) error {
	ownerRefs := []interface{}{
		map[string]interface{}{
			"apiVersion":         istioAPIVersion,
			"blockOwnerDeletion": true,
			"controller":         true,
			"kind":               "istio-installer",
			"name":               d.componentObject.GetName(),
			"uid":                string(d.componentObject.GetUID()),
		},
	}

	for _, o := range manifestObjects.Items {
		if err := o.SetNestedField(ownerRefs, "metadata", "ownerReferences"); err != nil {
			return err
		}
	}
	return nil
}

func cRDKindObjects(objects *manifest.Objects) *manifest.Objects {
	var ret *manifest.Objects
	for _, o := range objects.Items {
		if o.Kind == "CustomResourceDefinition" {
			ret.Items = append(ret.Items, o)
		}
	}
	return ret
}

func (d *DeploymentComponent) waitForCRDs(objects *manifest.Objects) error {
	log.Info("Waiting for CRDs to be applied.")
	cs, err := apiextensionsclient.NewForConfig(d.config)
	if err != nil {
		return err
	}

	var crdNames []string
	for _, o := range cRDKindObjects(objects).Items {
		crdNames = append(crdNames, o.Name)
	}

	errPoll := wait.Poll(cRDPollInterval, cRDPollTimeout, func() (bool, error) {
	descriptor:
		for _, crdName := range crdNames {
			crd, errGet := cs.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crdName, metav1.GetOptions{})
			if errGet != nil {
				return false, errGet
			}
			for _, cond := range crd.Status.Conditions {
				switch cond.Type {
				case apiextensionsv1beta1.Established:
					if cond.Status == apiextensionsv1beta1.ConditionTrue {
						log.Infof("established CRD %q", crdName)
						continue descriptor
					}
				case apiextensionsv1beta1.NamesAccepted:
					if cond.Status == apiextensionsv1beta1.ConditionFalse {
						log.Warnf("name conflict: %v", cond.Reason)
					}
				}
			}
			log.Infof("missing status condition for %q", crdName)
			return false, nil
		}
		return true, nil
	})

	if errPoll != nil {
		log.Error("failed to verify CRD creation")
		return errPoll
	}

	log.Info("CRDs applied.")
	return nil
}

// +k8s:deepcopy-gen=true
type DeploymentStatus struct {
	Healthy    bool                         `json:"healthy"`
	Deployment *corev1.ObjectReference      `json:"deployment,omitempty"`
	Conditions []appsv1.DeploymentCondition `json:"conditions,omitempty"`
}

// +k8s:deepcopy-gen=true
type ServiceStatus struct {
	Healthy bool                    `json:"healthy"`
	Service *corev1.ObjectReference `json:"service,omitempty"`
}

func (d *DeploymentComponent) buildServiceStatus(ctx context.Context, namespace, name string) (ServiceStatus, error) {
	ss := ServiceStatus{}

	svcName := client.ObjectKey{namespace, name}
	svc := &corev1.Service{}
	err := d.client.Get(ctx, svcName, svc)
	if err != nil {
		return ss, fmt.Errorf("get service %s: %v", svcName.String(), err)
	}

	ss.Service = &corev1.ObjectReference{
		Kind:       "Service",
		APIVersion: "v1",
		Namespace:  svc.Namespace,
		Name:       svc.Name,
		UID:        svc.UID,
	}

	ss.Healthy = true

	return ss, nil
}

func (d *DeploymentComponent) buildDeploymentStatus(ctx context.Context, namespace, name string) (DeploymentStatus, error) {
	ds := DeploymentStatus{}

	dep := &appsv1.Deployment{}
	err := d.client.Get(ctx, client.ObjectKey{namespace, name}, dep)
	if err != nil {
		return ds, fmt.Errorf("error reading deployment %s/%s: %v", namespace, name, err)
	}

	ds.Deployment = &corev1.ObjectReference{
		Kind:       "Deployment",
		APIVersion: "apps/v1",
		Namespace:  dep.Namespace,
		Name:       dep.Name,
		UID:        dep.UID,
	}

	for _, cond := range dep.Status.Conditions {
		ds.Conditions = append(ds.Conditions, cond)
		// TODO: This is basing the entire health of the Deployment from this condition
		// Is that enough?
		if cond.Type == appsv1.DeploymentAvailable {
			ds.Healthy = cond.Status == corev1.ConditionTrue
		}
	}

	return ds, nil
}

func defaultObjectOrder() func(o *manifest.Object) int {
	return func(o *manifest.Object) int {
		gk := o.Group + "/" + o.Kind
		switch gk {
		// Create CRDs asap - both because they are slow and because we will likely create instances of them soon
		case "apiextensions.k8s.io/CustomResourceDefinition":
			return -1000

			// We need to create ServiceAccounts, Roles before we bind them with a RoleBinding
		case "/ServiceAccount", "rbac.authorization.k8s.io/ClusterRole":
			return 1
		case "rbac.authorization.k8s.io/ClusterRoleBinding":
			return 2

			// Pods might need configmap or secrets - avoid backoff by creating them first
		case "/ConfigMap", "/Secrets":
			return 100

			// Create the pods after we've created other things they might be waiting for
		case "extensions/Deployment", "app/Deployment":
			return 1000

			// Autoscalers typically act on a deployment
		case "autoscaling/HorizontalPodAutoscaler":
			return 1001

			// Create services late - after pods have been started
		case "/Service":
			return 10000

		default:
			log.Error("unknown group / kind")
			return 1000
		}
	}
}

/*
func (d *DeploymentComponent) ensureWatches(ctx context.Context, name types.NamespacedName, objects *manifest.Objects) error {
	log := klog.FromContext(ctx)
	log.Info().Message("ensuring watches")

	labelSelector := fmt.Sprintf("component=%s", fields.EscapeValue(d.componentName))

	notify := metav1.ObjectMeta{Name: name.Name, Namespace: name.Namespace}
	filter := metav1.ListOptions{LabelSelector: labelSelector}

	for _, gvk := range uniqueGroupVersionKind(objects) {
		err := r.options.watch.Add(gvk, filter, notify)
		if err != nil {
			log.Error().Err(err).Tag("GroupVersionKind", gvk.String()).Message("adding watch")
			continue
		}
	}

	return nil
}

type dynamicWatch struct {
	clientPool deprecated_dynamic.ClientPool
	restMapper meta.RESTMapper
	events     chan event.GenericEvent
}

func newDynamicWatch(config rest.Config) (*dynamicWatch, error) {
	dw := &dynamicWatch{events: make(chan event.GenericEvent)}

	restMapper, err := apiutil.NewDiscoveryRESTMapper(&config)
	if err != nil {
		return nil, err
	}

	dw.restMapper = restMapper
	dw.clientPool = deprecated_dynamic.NewClientPool(&config, dw.restMapper, dynamic.LegacyAPIPathResolverFunc)
	return dw, nil
}

func (d *DeploymentComponent) WatchAllDeployedObjects() error {

	watch, err := newDynamicWatch(d.config)
	if err != nil {
		return fmt.Errorf("creating dynamic watch: %v", err)
	}
	src := &source.Channel{Source: watch.events}
	// Inject a stop channel that will never close. The controller does not have a concept of
	// shutdown, so there is no oppritunity to stop the watch.
	stopCh := make(chan struct{})
	src.InjectStopChannel(stopCh)
	if err := controller.Watch(src, &handler.EnqueueRequestForObject{}); err != nil {
		return fmt.Errorf("setting up dynamic watch on the controller: %v", err)
	}

	d.watch = watch

	return nil
}
*/
