package feature

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
	"github.com/ostromart/istio-installer/pkg/component/component"
	"github.com/ostromart/istio-installer/pkg/util"
)

const (
	TrafficManagementFeatureName = "TrafficManagement"
	PolicyFeatureName            = "Policy"
	TelemetryFeatureName         = "Telemetry"
	SecurityFeatureName          = "Security"
	ConfigManagementFeatureName  = "ConfigManagement"
	AutoInjectionFeatureName     = "AutoInjection"
)

type Feature interface {
	Run() error
	RenderManifest() (string, util.Errors)
}

type FeatureOptions struct {
	InstallSpec      *v1alpha1.IstioControlPlaneSpec
	Dirs             component.ComponentDirLayout
	GlobalValuesFile string
	HelmChartName    string
	HelmChartDir     string
}

type CommonFeatureFields struct {
	FeatureOptions
	components []component.Component
}

type TrafficManagementFeature struct {
	CommonFeatureFields
}

func newComponentOptions(cff *CommonFeatureFields, featureName, componentName string) *component.ComponentOptions {
	return &component.ComponentOptions{
		InstallSpec:      cff.InstallSpec,
		FeatureName:      featureName,
		GlobalValuesFile: cff.GlobalValuesFile,
		HelmChartName:    cff.HelmChartName,
		HelmChartDir:     filepath.Join(cff.HelmChartDir, cff.Dirs[componentName]),
	}
}

func NewTrafficManagementFeature(opts *FeatureOptions) *TrafficManagementFeature {
	cff := &CommonFeatureFields{
		FeatureOptions: *opts,
	}
	cff.components = []component.Component{
		component.NewPilotComponent(newComponentOptions(cff, TrafficManagementFeatureName, component.PilotComponentName)),
		component.NewSidecarInjectorComponent(newComponentOptions(cff, TrafficManagementFeatureName, component.SidecarInjectorComponentName)),
	}

	return &TrafficManagementFeature{
		CommonFeatureFields: *cff,
	}
}

func (f *TrafficManagementFeature) Run() error {
	return runComponents(f.components)
}

func (f *TrafficManagementFeature) RenderManifest() (string, util.Errors) {
	fmt.Printf("Render TrafficManagementFeature\n")
	return renderComponents(f.components)
}

type SecurityFeature struct {
	CommonFeatureFields
}

func (f *SecurityFeature) Run() error {
	return runComponents(f.components)
}

func NewSecurityFeature(opts *FeatureOptions) *SecurityFeature {
	cff := &CommonFeatureFields{
		FeatureOptions: *opts,
	}
	cff.components = []component.Component{
		component.NewCitadelComponent(newComponentOptions(cff, SecurityFeatureName, component.CitadelComponentName)),
		component.NewCertManagerComponent(newComponentOptions(cff, SecurityFeatureName, component.CertManagerComponentName)),
		component.NewNodeAgentComponent(newComponentOptions(cff, SecurityFeatureName, component.NodeAgentComponentName)),
	}
	return &SecurityFeature{
		CommonFeatureFields: *cff,
	}
}

func (f *SecurityFeature) RenderManifest() (string, util.Errors) {
	return renderComponents(f.components)
}

type PolicyFeature struct {
	CommonFeatureFields
}

func (f *PolicyFeature) Run() error {
	return runComponents(f.components)
}

func NewPolicyFeature(opts *FeatureOptions) *PolicyFeature {
	cff := &CommonFeatureFields{
		FeatureOptions: *opts,
	}
	cff.components = []component.Component{
		component.NewPolicyComponent(newComponentOptions(cff, PolicyFeatureName, component.PolicyComponentName)),
	}
	return &PolicyFeature{
		CommonFeatureFields: *cff,
	}
}

func (f *PolicyFeature) RenderManifest() (string, util.Errors) {
	return renderComponents(f.components)
}

type TelemetryFeature struct {
	CommonFeatureFields
}

func (f *TelemetryFeature) Run() error {
	return runComponents(f.components)
}

func NewTelemetryFeature(opts *FeatureOptions) *TelemetryFeature {
	cff := &CommonFeatureFields{
		FeatureOptions: *opts,
	}
	cff.components = []component.Component{
		component.NewTelemetryComponent(newComponentOptions(cff, TelemetryFeatureName, component.TelemetryComponentName)),
	}
	return &TelemetryFeature{
		CommonFeatureFields: *cff,
	}
}

func (f *TelemetryFeature) RenderManifest() (string, util.Errors) {
	return renderComponents(f.components)
}

type ConfigManagementFeature struct {
	CommonFeatureFields
}

func (f *ConfigManagementFeature) Run() error {
	return runComponents(f.components)
}

func NewConfigManagementFeature(opts *FeatureOptions) *ConfigManagementFeature {
	cff := &CommonFeatureFields{
		FeatureOptions: *opts,
	}
	cff.components = []component.Component{
		component.NewGalleyComponent(newComponentOptions(cff, ConfigManagementFeatureName, component.GalleyComponentName)),
	}
	return &ConfigManagementFeature{
		CommonFeatureFields: *cff,
	}
}

func (f *ConfigManagementFeature) RenderManifest() (string, util.Errors) {
	return renderComponents(f.components)
}

type AutoInjectionFeature struct {
	CommonFeatureFields
}

func (f *AutoInjectionFeature) Run() error {
	return runComponents(f.components)
}

func NewAutoInjectionFeature(opts *FeatureOptions) *AutoInjectionFeature {
	cff := &CommonFeatureFields{
		FeatureOptions: *opts,
	}
	cff.components = []component.Component{
		component.NewSidecarInjectorComponent(newComponentOptions(cff, AutoInjectionFeatureName, component.SidecarInjectorComponentName)),
	}
	return &AutoInjectionFeature{
		CommonFeatureFields: *cff,
	}
}

func (f *AutoInjectionFeature) RenderManifest() (string, util.Errors) {
	return renderComponents(f.components)
}

func runComponents(cs []component.Component) error {
	for _, c := range cs {
		if err := c.Run(); err != nil {
			return err
		}
	}
	return nil
}

func renderComponents(cs []component.Component) (manifest string, errsOut util.Errors) {
	var sb strings.Builder
	for _, c := range cs {
		s, err := c.RenderManifest()
		errsOut = util.AppendErr(errsOut, err)
		_, err = sb.WriteString(s)
		errsOut = util.AppendErr(errsOut, err)
	}
	if len(errsOut) > 0 {
		return "", errsOut
	}
	return sb.String(), nil
}

func checkOptions(fo *FeatureOptions) error {
	if fo.InstallSpec == nil {
		return fmt.Errorf("InstallSpec must be set")
	}
	if fo.HelmChartName == "" {
		return fmt.Errorf("HelmChartName must be set")
	}
	if fo.HelmChartDir == "" {
		return fmt.Errorf("HelmChartDir must be set")
	}
	if fo.Dirs == nil {
		return fmt.Errorf("Dirs must be set")
	}
	return nil
}
