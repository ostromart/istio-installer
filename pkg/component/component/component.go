package component

import (
	protobuf "github.com/gogo/protobuf/types"
	"github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
	"github.com/ostromart/istio-installer/pkg/helm"
	"github.com/ostromart/istio-installer/pkg/patch"
	"gopkg.in/yaml.v2"
)

const (
	IstioBaseComponentName       = "crds"
	PilotComponentName           = "pilot"
	GalleyComponentName          = "galley"
	SidecarInjectorComponentName = "sidecar-injector"
	PolicyComponentName          = "policy"
	TelemetryComponentName       = "telemetry"
	CitadelComponentName         = "citadel"
	CertManagerComponentName     = "cert-manager"
	NodeAgentComponentName       = "node-agent"
	IngressComponentName         = "ingress"
	EgressComponentName          = "egress"

	componentDisabledStr = " component is disabled."
	yamlCommentStr       = "# "
)

type ComponentDirLayout map[string]string

var (
	V12 = ComponentDirLayout{
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
	FeatureEnabled   bool
	FeatureNamespace string
	HelmChartName    string
	HelmChartDir     string
	InstallSpec      *v1alpha1.IstioControlPlaneSpec
}

type Component interface {
	RenderManifest() (string, error)
}

type CommonComponentFields struct {
	ComponentOptions
	enabled   bool
	namespace string
	name      string
}

type PilotComponent struct {
	CommonComponentFields
}

func (c *PilotComponent) RenderManifest() (string, error) {
	if !c.enabled {
		return yamlCommentStr + c.name + componentDisabledStr, nil
	}
	renderer, err := helm.NewHelmTemplateRenderer(c.HelmChartDir, c.HelmChartName, c.namespace)
	if err != nil {
		return "", err
	}

	var baseValues []byte
	if c.hasValuesOverrides() {
		baseValues, err = yaml.Marshal(c.InstallSpec.TrafficManagement.Pilot.Common.ValuesOverrides)
	}
	if err != nil {
		return "", err
	}
	// TODO: overlay values.

	baseYAML, err := renderer.Render(string(baseValues))
	if err != nil {
		return "", err
	}

	patched := baseYAML
	if c.hasK8sOverrides() {
		patched, err = patch.PatchYAMLManifest(baseYAML, c.namespace, c.InstallSpec.TrafficManagement.Pilot.Common.K8S.Overlays)
		if err != nil {
			return "", err
		}
	}
	return patched, nil
}

func NewPilotComponent(opts *ComponentOptions) *PilotComponent {
	ret := &PilotComponent{
		CommonComponentFields{
			ComponentOptions: *opts,
			name:             PilotComponentName,
		},
	}
	if opts.InstallSpec.TrafficManagement.Pilot != nil &&
		opts.InstallSpec.TrafficManagement.Pilot.Common != nil {
		ret.CommonComponentFields.enabled = withOverrideBool(opts.FeatureEnabled, opts.InstallSpec.TrafficManagement.Pilot.Common.Enabled)
		ret.CommonComponentFields.namespace = withOverrideString(opts.FeatureNamespace, opts.InstallSpec.TrafficManagement.Pilot.Common.Namespace)
	}
	return ret
}

type ProxyComponent struct {
}

func (c *ProxyComponent) RenderManifest() (string, error) {
	return "", nil
}

func NewProxyComponent(opts *ComponentOptions) *ProxyComponent {
	return nil
}

type CitadelComponent struct {
}

func (c *CitadelComponent) RenderManifest() (string, error) {
	return "", nil
}

func NewCitadelComponent(opts *ComponentOptions) *CitadelComponent {
	return nil
}

type CertManagerComponent struct {
}

func (c *CertManagerComponent) RenderManifest() (string, error) {
	return "", nil
}

func NewCertManagerComponent(opts *ComponentOptions) *CertManagerComponent {
	return nil
}

type NodeAgentComponent struct {
}

func (c *NodeAgentComponent) RenderManifest() (string, error) {
	return "", nil
}

func NewNodeAgentComponent(opts *ComponentOptions) *NodeAgentComponent {
	return nil
}

type PolicyComponent struct {
}

func (c *PolicyComponent) RenderManifest() (string, error) {
	return "", nil
}

func NewPolicyComponent(opts *ComponentOptions) *PolicyComponent {
	return nil
}

type TelemetryComponent struct {
}

func (c *TelemetryComponent) RenderManifest() (string, error) {
	return "", nil
}

func NewTelemetryComponent(opts *ComponentOptions) *TelemetryComponent {
	return nil
}

type GalleyComponent struct {
}

func (c *GalleyComponent) RenderManifest() (string, error) {
	return "", nil
}

func NewGalleyComponent(opts *ComponentOptions) *GalleyComponent {
	return nil
}

type SidecarInjectorComponent struct {
}

func (c *SidecarInjectorComponent) RenderManifest() (string, error) {
	return "", nil
}

func NewSidecarInjectorComponent(opts *ComponentOptions) *SidecarInjectorComponent {
	return nil
}


func withOverrideBool(base bool, override *protobuf.BoolValue) bool {
	if override == nil {
		return base
	}
	return override.Value
}

func withOverrideString(base string, override string) string {
	if override == "" {
		return base
	}
	return override
}
func (c *PilotComponent) hasValuesOverrides() bool {
	return c.InstallSpec.TrafficManagement.Pilot != nil &&
		c.InstallSpec.TrafficManagement.Pilot.Common != nil &&
		c.InstallSpec.TrafficManagement.Pilot.Common.ValuesOverrides != nil
}

func (c *PilotComponent) hasK8sOverrides() bool {
	return c.InstallSpec.TrafficManagement.Pilot != nil &&
		c.InstallSpec.TrafficManagement.Pilot.Common != nil &&
		c.InstallSpec.TrafficManagement.Pilot.Common.K8S != nil &&
		c.InstallSpec.TrafficManagement.Pilot.Common.K8S.Overlays != nil
}


