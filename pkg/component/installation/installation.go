package installation

import (
	"fmt"
	"os"
	"path/filepath"
	"io/ioutil"
	"sync"
	"context"
	"strings"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"github.com/ostromart/istio-installer/pkg/component/component"
	installerv1alpha1 "github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
	"istio.io/istio/pkg/log"
)

const (
	// rootComponentName is the Istio root component name.
	rootComponentName = "istio"
	// relativeSubchartDirPath is the relative path from the chart path base dir to the subcharts dir.
	relativeSubchartDirPath = "../subcharts"

	// Names of components, corresponding to helm subchart dir names.
	certmanagerName            = "certmanager"
	gallyeName                 = "galley"
	gatewaysName               = "gateways"
	grafanaName                = "grafana"
	ingressName                = "ingress"
	istiocorednsName           = "istiocoredns"
	kialiName                  = "kiali"
	mixerName                  = "mixer"
	nodeagentName              = "nodeagent"
	pilotName                  = "pilot"
	prometheusName             = "prometheus"
	securityName               = "security"
	servicegraphName           = "servicegraph"
	sidecarInjectorWebhookName = "sidecarInjectorWebhook"
	telemetryGatewayName       = "telemetry-gateway"
	tracingName                = "tracing"
	cNIName                    = "istio-cni"
)

// IstioInstallation is an installation of Istio comprising a pointer to templates and two sources of config:
// 1. InstallerSpec, which specifies whether top level functional component are enabled, and any overrides to
//    baseValues and k8s resources.
// 2. Values, which is the default baseValues manifest.
// From these sources, component manifests are created and may be rendered, applied and optionally reconciled.
type IstioInstallation struct {
	// helmChartPath is the path of the root of helm charts in the local filesystem.
	helmChartPath string
	// installerSpec is the Istio install CRD.
	installerSpec *installerv1alpha1.InstallerSpec
	// baseValues is a default tree of baseValues from the baseValues manifest.
	baseValues      map[string]interface{}
	installed       map[string]*component.DeploymentComponent
	rootComponent   *component.DeploymentComponent
	installTree     deploymentTree
	config          map[string]*component.ComponentDeploymentConfig
	nameToNamespace map[string]string
	// k8sClient is the kubernetes client.
	k8sClient client.Client
	// k8sConfig is the config for kubernetes client.
	k8sConfig *rest.Config
	// k8sClientset is the kubernetes Clientset.
	k8sClientset *kubernetes.Clientset
}

type deploymentTree map[*component.DeploymentComponent]interface{}

func NewIstioInstallation(helmChartPath string, k8sClient client.Client, k8sConfig *rest.Config, k8sClientset *kubernetes.Clientset, baseValues map[string]interface{}) *IstioInstallation {
	return &IstioInstallation{
		helmChartPath: helmChartPath,
		baseValues:    baseValues,
		k8sClient:     k8sClient,
		k8sConfig:     k8sConfig,
		k8sClientset:  k8sClientset,

		installed:       make(map[string]*component.DeploymentComponent),
		installTree:     make(deploymentTree),
		config:          make(map[string]*component.ComponentDeploymentConfig),
		nameToNamespace: make(map[string]string),
	}
}

func (cm *IstioInstallation) Build(installerSpec *installerv1alpha1.InstallerSpec) {
	cm.installerSpec = installerSpec
	cm.parse()
	cm.createComponents()
}

func (cm *IstioInstallation) RenderToDir(outputDir string) error {
	log.Infof("Rendering manifests to output dir %s", outputDir)
	return cm.renderRecursive(cm.installTree, outputDir)
}

func (cm *IstioInstallation) renderRecursive(tree deploymentTree, outputDir string) error {
	for k, v := range tree {
		log.Infof("Rendering: %s", k.Name())
		dirName := filepath.Join(outputDir, k.Name())
		if err := os.MkdirAll(dirName, os.ModePerm); err != nil {
			return fmt.Errorf("could not create directory %s; %s", outputDir, err)
		}
		str, err := k.Render(cm.installerSpec)
		if err != nil {
			return fmt.Errorf("could not generate config; %s", err)
		}
		fname := filepath.Join(dirName, k.Name()) + ".yaml"
		log.Infof("Writing manifest to %s", fname)
		if err = ioutil.WriteFile(fname, []byte(str), 0644); err != nil {
			return fmt.Errorf("could not write manifest config; %s", err)
		}

		kt, ok := v.(deploymentTree)
		if !ok {
			// Leaf
			return nil
		}
		if err := cm.renderRecursive(kt, dirName); err != nil {
			return err
		}
	}
	return nil
}

func (cm *IstioInstallation) ApplyOnce(ctx context.Context) error {
	log.Info("Rendering and applying manifests.")
	var err error
	var wg sync.WaitGroup

	// Start everything in parallel and let dependencies take care of any sequencing.
	for n, c := range cm.installed {
		log.Infof("Rendering: %s", n)
		wg.Add(1)
		go func() {
			err2 := c.ApplyOnce(ctx, cm.installerSpec)
			if err2 != nil {
				log.Errorf("could not generate config; %s", err2)
				err = err2
				return
			}
			wg.Done()
		}()
	}
	wg.Wait()
	return err
}

func (cm *IstioInstallation) RunApplyLoop(ctx context.Context) error {
	cm.Build(cm.installerSpec)
	for _, c := range cm.installed {
		if err := c.RunApplyLoop(context.Background()); err != nil {
			return fmt.Errorf("apply loop failed: %s", err)
		}
	}
	return nil
}

func (cm *IstioInstallation) createComponents() {
	log.Info("Creating components.")
	cm.rootComponent = component.NewComponentDeployment(&component.ComponentDeploymentConfig{
		ComponentName:        rootComponentName,
		Namespace:            cm.installerSpec.DefaultNamespace,
		HelmChartBaseDirPath: cm.helmChartPath,
		RelativeChartPath:    "..",
		Dependencies:         nil,
	})
	cm.installed[rootComponentName] = cm.rootComponent

	for k, v := range cm.nameToNamespace {
		log.Infof("  Creating %s", k)
		cm.installed[k] = component.NewComponentDeployment(&component.ComponentDeploymentConfig{
			ComponentName:        k,
			Namespace:            v,
			HelmChartBaseDirPath: cm.helmChartPath,
			RelativeChartPath:    relativeSubchartDirPath,
			Dependencies: []*component.Dependency{
				{
					DeploymentComponent: cm.rootComponent,
					Version:             "",
				},
			},
		})
	}
	cm.buildInstallTree()
}

type componentToListMap map[*component.DeploymentComponent][]*component.DeploymentComponent

func (cm *IstioInstallation) buildInstallTree() {
	log.Info("Building install tree.")

	// Create a map of all immediate children of all components.
	children := make(map[*component.DeploymentComponent][]*component.DeploymentComponent)
	for _, dComp := range cm.installed {
		for _, d := range dComp.Dependencies() {
			dc := d.DeploymentComponent
			children[dc] = append(children[dc], dComp)
		}
	}

	// Starting with root, recursively insert each first level child into each node.
	insertChildrenRecursive(cm.installed[rootComponentName], cm.installTree, children)

	log.Infof("Built install dependency tree: \n%s", cm.installTreeString())
}

func insertChildrenRecursive(dc *component.DeploymentComponent, tree deploymentTree, children componentToListMap) {
	tree[dc] = make(deploymentTree)
	for _, child := range children[dc] {
		insertChildrenRecursive(child, tree[dc].(deploymentTree), children)
	}
}

func (cm *IstioInstallation) installTreeString() string {
	var sb strings.Builder
	cm.buildInstallTreeString(cm.installed[rootComponentName], "", &sb)
	return sb.String()
}

func (cm *IstioInstallation) buildInstallTreeString(dc *component.DeploymentComponent, prefix string, sb *strings.Builder) {
	sb.WriteString(prefix + dc.Name() + "\n")
	if _, ok := cm.installTree[dc].(deploymentTree); !ok {
		return
	}
	for k := range cm.installTree[dc].(deploymentTree) {
		cm.buildInstallTreeString(k, prefix+"  ", sb)
	}
}

func (cm *IstioInstallation) parse() {
	if cm.installerSpec.TrafficManagement.Enabled {
		cm.nameToNamespace[pilotName] = withDefault(cm.installerSpec.TrafficManagement.Namespace, cm.installerSpec.DefaultNamespace)
	}
}

func (cm *IstioInstallation) getRootComponent() *component.DeploymentComponent {
	var ret *component.DeploymentComponent

	for k := range cm.installTree {
		return k
	}
	return ret
}

func withDefault(val, dflt string) string {
	if val == "" {
		return dflt
	}
	return val
}
