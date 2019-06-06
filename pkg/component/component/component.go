package component

import (
	"fmt"
	"reflect"
	"strings"

	"istio.io/pkg/log"

	"github.com/ostromart/istio-installer/pkg/util"

	protobuf "github.com/gogo/protobuf/types"
	"github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
	"github.com/ostromart/istio-installer/pkg/helm"
	"github.com/ostromart/istio-installer/pkg/patch"
	"gopkg.in/yaml.v2"
)

const (
	IstioBaseComponentName       = "crds"
	PilotComponentName           = "Pilot"
	GalleyComponentName          = "Galley"
	SidecarInjectorComponentName = "SidecarInjector"
	PolicyComponentName          = "Policy"
	TelemetryComponentName       = "Telemetry"
	CitadelComponentName         = "Citadel"
	CertManagerComponentName     = "CertManager"
	NodeAgentComponentName       = "NodeAgent"
	IngressComponentName         = "Ingress"
	EgressComponentName          = "Egress"

	componentDisabledStr = " component is disabled."
	yamlCommentStr       = "# "
)

type ComponentDirLayout map[string]string

var (
	V12DirLayout = ComponentDirLayout{
		PilotComponentName:           "istio-control/istio-discovery",
		GalleyComponentName:          "istio-control/istio-config",
		SidecarInjectorComponentName: "istio-control/istio-autoinject",
		PolicyComponentName:          "istio-policy",
		TelemetryComponentName:       "istio-telemetry",
		CitadelComponentName:         "security/citadel",
		NodeAgentComponentName:       "security/nodeagent",
		CertManagerComponentName:     "security/certmanager",
		IngressComponentName:         "gateways/istio-ingress",
		EgressComponentName:          "gateways/istio-egress",
	}
)

type ComponentOptions struct {
	FeatureName      string
	HelmChartName    string
	HelmChartDir     string
	GlobalValuesFile string
	InstallSpec      *v1alpha1.IstioControlPlaneSpec
}

type Component interface {
	Run() error
	RenderManifest() (string, error)
}

type CommonComponentFields struct {
	*ComponentOptions
	enabled   bool
	namespace string
	name      string
	renderer  helm.TemplateRenderer
	started   bool
}

type PilotComponent struct {
	*CommonComponentFields
}

func (c *PilotComponent) Run() error {
	return runComponent(c.CommonComponentFields)
}

func (c *PilotComponent) RenderManifest() (string, error) {
	fmt.Printf("Render PilotComponent\n")
	if !c.started {
		return "", fmt.Errorf("component %s not started in RenderManifest", c.name)
	}
	return renderManifest(c.CommonComponentFields)
}

func NewPilotComponent(opts *ComponentOptions) *PilotComponent {
	ret := &PilotComponent{
		&CommonComponentFields{
			ComponentOptions: opts,
			name:             PilotComponentName,
		},
	}
	return ret
}

func disabledYAMLStr(componentName string) string {
	return yamlCommentStr + componentName + componentDisabledStr
}

func patchTree(root, patch map[string]interface{}) {
	// TODO: implement
}

func runComponent(c *CommonComponentFields) error {
	r, err := createHelmRenderer(c)
	if err != nil {
		return err
	}
	if err := r.Run(); err != nil {
		return err
	}
	c.renderer = r
	c.started = true
	return nil
}

func isComponentEnabled(featureName, componentName string, installSpec *v1alpha1.IstioControlPlaneSpec) bool {
	featureNodeI, err := GetFromStructPath(installSpec, featureName+".Enabled")
	if err != nil {
		log.Error(err.Error())
		return false
	}
	if featureNodeI == nil {
		return false
	}
	featureNode, ok := featureNodeI.(*protobuf.BoolValue)
	if !ok {
		log.Errorf("feature %s enabled has bad type %T, expect *protobuf.BoolValue", featureNodeI)
	}
	if featureNode == nil {
		return false
	}
	if featureNode.Value == false {
		return false
	}

	componentNodeI, err := GetFromStructPath(installSpec, featureName+".Components."+componentName+".Enabled")
	if err != nil {
		log.Error(err.Error())
		return false
	}
	if componentNodeI == nil {
		return true
	}
	componentNode, ok := componentNodeI.(*protobuf.BoolValue)
	if !ok {
		log.Errorf("component %s enabled has bad type %T, expect *protobuf.BoolValue", componentNodeI)
	}
	if componentNode == nil {
		return false
	}
	return componentNode.Value
}

func renderManifest(c *CommonComponentFields) (string, error) {
	if !isComponentEnabled(c.FeatureName, c.name, c.InstallSpec) {
		fmt.Printf("disabled\n")
		return disabledYAMLStr(c.name), nil
	}

	var vals, valsUnvalidated map[string]interface{}
	err := SetFromPath(c.ComponentOptions.InstallSpec, "TrafficManagement.Components."+c.name+".Common.ValuesOverrides", vals)
	if err != nil {
		return "", err
	}
	err = SetFromPath(c.ComponentOptions.InstallSpec, "TrafficManagement.Components."+c.name+".Common.UnvalidatedValuesOverrides", valsUnvalidated)
	if err != nil {
		return "", err
	}

	patchTree(vals, valsUnvalidated)

	valsYAML, err := yaml.Marshal(vals)
	if err != nil {
		return "", err
	}

	my, err := c.renderer.Render(string(valsYAML))
	if err != nil {
		return "", err
	}
	my += helm.YAMLSeparator + "\n"

	var overlays []*v1alpha1.K8SObjectOverlay
	err = SetFromPath(c.InstallSpec, "TrafficManagement.Components."+c.name+".Common.K8s.Overlays", overlays)
	if err != nil {
		return "", err
	}

	return patch.PatchYAMLManifest(my, c.namespace, overlays)
}

func createHelmRenderer(c *CommonComponentFields) (helm.TemplateRenderer, error) {
	cp := c.InstallSpec.CustomPackagePath
	switch {
	case cp == "":
		return nil, fmt.Errorf("compiled in CustomPackagePath not yet supported")
	case isFilePath(cp):
		return helm.NewFileTemplateRenderer(c.GlobalValuesFile, c.HelmChartDir, c.name, c.namespace), nil
	default:
	}
	return nil, fmt.Errorf("unsupported CustomPackagePath %s", cp)
}

func isFilePath(path string) bool {
	return strings.HasPrefix(path, "file://")
}

type ProxyComponent struct {
}

func (c *ProxyComponent) Run() error {
	return nil
}

func (c *ProxyComponent) RenderManifest() (string, error) {
	return "", nil
}

func NewProxyComponent(opts *ComponentOptions) *ProxyComponent {
	return nil
}

type CitadelComponent struct {
}

func (c *CitadelComponent) Run() error {
	return nil
}

func (c *CitadelComponent) RenderManifest() (string, error) {
	return "", nil
}

func NewCitadelComponent(opts *ComponentOptions) *CitadelComponent {
	return nil
}

type CertManagerComponent struct {
}

func (c *CertManagerComponent) Run() error {
	return nil
}

func (c *CertManagerComponent) RenderManifest() (string, error) {
	return "", nil
}

func NewCertManagerComponent(opts *ComponentOptions) *CertManagerComponent {
	return nil
}

type NodeAgentComponent struct {
}

func (c *NodeAgentComponent) Run() error {
	return nil
}

func (c *NodeAgentComponent) RenderManifest() (string, error) {
	return "", nil
}

func NewNodeAgentComponent(opts *ComponentOptions) *NodeAgentComponent {
	return nil
}

type PolicyComponent struct {
}

func (c *PolicyComponent) Run() error {
	return nil
}

func (c *PolicyComponent) RenderManifest() (string, error) {
	return "", nil
}

func NewPolicyComponent(opts *ComponentOptions) *PolicyComponent {
	return nil
}

type TelemetryComponent struct {
}

func (c *TelemetryComponent) Run() error {
	return nil
}

func (c *TelemetryComponent) RenderManifest() (string, error) {
	return "", nil
}

func NewTelemetryComponent(opts *ComponentOptions) *TelemetryComponent {
	return nil
}

type GalleyComponent struct {
}

func (c *GalleyComponent) Run() error {
	return nil
}

func (c *GalleyComponent) RenderManifest() (string, error) {
	return "", nil
}

func NewGalleyComponent(opts *ComponentOptions) *GalleyComponent {
	return nil
}

type SidecarInjectorComponent struct {
}

func (c *SidecarInjectorComponent) Run() error {
	return nil
}

func (c *SidecarInjectorComponent) RenderManifest() (string, error) {
	return "", nil
}

func NewSidecarInjectorComponent(opts *ComponentOptions) *SidecarInjectorComponent {
	return nil
}

func SetFromPath(node interface{}, path string, out interface{}) error {
	val, err := GetFromStructPath(node, path)
	if err != nil {
		return err
	}
	if util.IsValueNil(val) {
		return nil
	}
	if reflect.TypeOf(val) != reflect.TypeOf(out) {
		return fmt.Errorf("SetFromPath from type %T != to type %T", val, out)
	}
	reflect.ValueOf(out).Set(reflect.ValueOf(val))
	return nil
}

func GetFromStructPath(node interface{}, path string) (interface{}, error) {
	return getFromStructPath(node, util.PathFromString(path))
}

func getFromStructPath(node interface{}, path util.Path) (interface{}, error) {
	if reflect.TypeOf(node).Kind() != reflect.Ptr {
		return nil, fmt.Errorf("GetFromStructPath path %s, expected struct ptr, got %T", path, node)
	}
	structElems := reflect.ValueOf(node).Elem()
	if reflect.TypeOf(structElems).Kind() != reflect.Struct {
		return nil, fmt.Errorf("GetFromStructPath path %s, expected struct ptr, got %T", path, node)
	}

	if len(path) == 0 {
		return node, nil
	}

	if util.IsNilOrInvalidValue(structElems) {
		return nil, nil
	}

	for i := 0; i < structElems.NumField(); i++ {
		fieldName := structElems.Type().Field(i).Name

		if fieldName != path[0] {
			continue
		}

		fv := structElems.Field(i)
		kind := structElems.Type().Field(i).Type.Kind()
		if kind != reflect.Ptr {
			return nil, fmt.Errorf("struct field %s is %T, expect struct ptr", fieldName, fv.Interface())
		}

		return getFromStructPath(fv.Interface(), path[1:])
	}

	return nil, fmt.Errorf("path %s not found from node type %T", path, node)
}
