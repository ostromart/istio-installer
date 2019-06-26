package manifest

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/ostromart/istio-installer/pkg/kubectlcmd"
	"github.com/ostromart/istio-installer/pkg/name"

	"istio.io/istio/pkg/kube"
	"istio.io/pkg/log"
)

const (
	// cRDPollInterval is how often the state of CRDs is polled when waiting for their creation.
	cRDPollInterval = 500 * time.Millisecond
	// cRDPollTimeout is the maximum wait time for all CRDs to be created.
	cRDPollTimeout = 60 * time.Second
)

type componentNameToListMap map[name.ComponentName][]name.ComponentName
type componentTree map[name.ComponentName]interface{}

var (
	componentDependencies = componentNameToListMap{
		name.IstioBaseComponentName: {
			name.PilotComponentName,
			name.PolicyComponentName,
			name.TelemetryComponentName,
			name.GalleyComponentName,
			name.CitadelComponentName,
			name.NodeAgentComponentName,
			name.CertManagerComponentName,
			name.SidecarInjectorComponentName,
			name.IngressComponentName,
			name.EgressComponentName,
		},
	}

	installTree      = make(componentTree)
	dependencyWaitCh = make(map[name.ComponentName]chan struct{})
	kubectl          = kubectlcmd.New()

	k8sRESTConfig *rest.Config
	k8sRESTClient *rest.RESTClient
)

func init() {
	buildInstallTree()
	for _, parent := range componentDependencies {
		for _, child := range parent {
			dependencyWaitCh[child] = make(chan struct{}, 1)
		}
	}

}

func RenderToDir(manifests name.ManifestMap, outputDir string) error {
	log.Infof("Component dependencies tree: \n%s", installTreeString())
	log.Infof("Rendering manifests to output dir %s", outputDir)
	return renderRecursive(manifests, installTree, outputDir)
}

func renderRecursive(manifests name.ManifestMap, installTree componentTree, outputDir string) error {
	for k, v := range installTree {
		componentName := string(k)
		ym := manifests[k]
		if ym == "" {
			log.Infof("Manifest for %s not found, skip.", componentName)
			continue
		}
		log.Infof("Rendering: %s", componentName)
		dirName := filepath.Join(outputDir, componentName)
		if err := os.MkdirAll(dirName, os.ModePerm); err != nil {
			return fmt.Errorf("could not create directory %s; %s", outputDir, err)
		}
		fname := filepath.Join(dirName, componentName) + ".yaml"
		log.Infof("Writing manifest to %s", fname)
		if err := ioutil.WriteFile(fname, []byte(ym), 0644); err != nil {
			return fmt.Errorf("could not write manifest config; %s", err)
		}

		kt, ok := v.(componentTree)
		if !ok {
			// Leaf
			return nil
		}
		if err := renderRecursive(manifests, kt, dirName); err != nil {
			return err
		}
	}
	return nil
}

func ApplyAll(manifests name.ManifestMap) error {
	log.Info("Apply manifests for these components:")
	for c := range manifests {
		log.Infof("- %s", c)
	}
	log.Infof("Component dependencies tree: \n%s", installTreeString())
	if err := initK8SRestClient(); err != nil {
		return err
	}
	applyRecursive(manifests)
	return nil
}

func applyRecursive(manifests name.ManifestMap) {
	var wg sync.WaitGroup
	for c, m := range manifests {
		c := c
		wg.Add(1)
		go func() {
			if s := dependencyWaitCh[c]; s != nil {
				log.Infof("%s is waiting on parent dependency...", c)
				<-s
				log.Infof("parent dependency for %s has unblocked", c)
			}
			if err := applyManifest(c, m); err != nil {
				log.Error(err.Error())
				return
			}
			// Signal all the components that depend on us.
			for _, ch := range componentDependencies[c] {
				log.Infof("signaling child %s", ch)
				dependencyWaitCh[ch] <- struct{}{}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func applyManifest(componentName name.ComponentName, manifestStr string) error {
	// REMOVE
	log.Infof("applyManifest for %s", componentName)
	return nil

	objects, err := ParseObjectsFromYAMLManifest(manifestStr)
	if err != nil {
		return err
	}
	if len(objects) == 0 {
		return nil
	}

	namespace := ""
	for _, o := range objects {
		o.AddLabels(map[string]string{"component": string(componentName)})
		if o.Namespace != "" {
			// All objects in a component have the same namespace.
			namespace = o.Namespace
		}
	}
	objects.Sort(defaultObjectOrder())

	extraArgs := []string{"--force", "--prune", "--selector", fmt.Sprintf("component=%s", componentName)}

	crdObjects := cRDKindObjects(objects)

	mcrd, err := crdObjects.JSONManifest()
	if err != nil {
		return err
	}
	ctx := context.Background()
	if err := kubectl.Apply(ctx, namespace, mcrd, extraArgs...); err != nil {
		return err
	}
	// Not all Istio components are robust to not yet created CRDs.
	if err := waitForCRDs(objects); err != nil {
		return err
	}

	m, err := objects.JSONManifest()
	if err != nil {
		return err
	}
	if err := kubectl.Apply(ctx, namespace, m, extraArgs...); err != nil {
		return err
	}

	return nil
}

func defaultObjectOrder() func(o *Object) int {
	return func(o *Object) int {
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

func cRDKindObjects(objects Objects) Objects {
	var ret Objects
	for _, o := range objects {
		if o.Kind == "CustomResourceDefinition" {
			ret = append(ret, o)
		}
	}
	return ret
}

func waitForCRDs(objects Objects) error {
	log.Info("Waiting for CRDs to be applied.")
	cs, err := apiextensionsclient.NewForConfig(k8sRESTConfig)
	if err != nil {
		return err
	}

	var crdNames []string
	for _, o := range cRDKindObjects(objects) {
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

func buildInstallTree() {
	// Starting with root, recursively insert each first level child into each node.
	insertChildrenRecursive(name.IstioBaseComponentName, installTree, componentDependencies)
}

func insertChildrenRecursive(componentName name.ComponentName, tree componentTree, children componentNameToListMap) {
	tree[componentName] = make(componentTree)
	for _, child := range children[componentName] {
		insertChildrenRecursive(child, tree[componentName].(componentTree), children)
	}
}

func installTreeString() string {
	var sb strings.Builder
	buildInstallTreeString(name.IstioBaseComponentName, "", &sb)
	return sb.String()
}

func buildInstallTreeString(componentName name.ComponentName, prefix string, sb *strings.Builder) {
	_, _ = sb.WriteString(prefix + string(componentName) + "\n")
	if _, ok := installTree[componentName].(componentTree); !ok {
		return
	}
	for k := range installTree[componentName].(componentTree) {
		buildInstallTreeString(k, prefix+"  ", sb)
	}
}

func initK8SRestClient() error {
	var err error
	if k8sRESTConfig != nil {
		return nil
	}
	k8sRESTConfig, err = defaultRestConfig("", "")
	if err != nil {
		return err
	}
	/*	k8sRESTConfig, err = rest.RESTClientFor(config)
		if err != nil {
			return nil, err
		}
		return &Client{config, restClient}, nil
	*/
	return nil
}

func defaultRestConfig(kubeconfig, configContext string) (*rest.Config, error) {
	config, err := kube.BuildClientConfig(kubeconfig, configContext)
	if err != nil {
		return nil, err
	}
	config.APIPath = "/api"
	config.GroupVersion = &v1.SchemeGroupVersion
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}
	return config, nil
}

// BuildClientConfig is a helper function that builds client config from a kubeconfig filepath.
// It overrides the current context with the one provided (empty to use default).
//
// This is a modified version of k8s.io/client-go/tools/clientcmd/BuildConfigFromFlags with the
// difference that it loads default configs if not running in-cluster.
func BuildClientConfig(kubeconfig, context string) (*rest.Config, error) {
	if kubeconfig != "" {
		info, err := os.Stat(kubeconfig)
		if err != nil || info.Size() == 0 {
			// If the specified kubeconfig doesn't exists / empty file / any other error
			// from file stat, fall back to default
			kubeconfig = ""
		}
	}

	//Config loading rules:
	// 1. kubeconfig if it not empty string
	// 2. In cluster config if running in-cluster
	// 3. Config(s) in KUBECONFIG environment variable
	// 4. Use $HOME/.kube/config
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	loadingRules.ExplicitPath = kubeconfig
	configOverrides := &clientcmd.ConfigOverrides{
		ClusterDefaults: clientcmd.ClusterDefaults,
		CurrentContext:  context,
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides).ClientConfig()
}
