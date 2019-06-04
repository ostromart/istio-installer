package feature

import (
	"fmt"
	"path/filepath"
	"strings"

	protobuf "github.com/gogo/protobuf/types"
	"github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
	"github.com/ostromart/istio-installer/pkg/component/component"
	"github.com/ostromart/istio-installer/pkg/util"
)


type Feature interface {
	RenderManifest() (string, util.Errors)
}

type FeatureOptions struct {
	InstallSpec   *v1alpha1.IstioControlPlaneSpec
	Dirs          component.ComponentDirLayout
	HelmChartName string
	HelmChartDir  string
}

type CommonFeatureFields struct {
	FeatureOptions
	enabled    bool
	namespace  string
	components []component.Component
}

type TrafficManagementFeature struct {
	CommonFeatureFields
}

func newComponentOptions(cff *CommonFeatureFields, componentName string) *component.ComponentOptions {
	return &component.ComponentOptions{
		InstallSpec:      cff.InstallSpec,
		FeatureEnabled:   cff.enabled,
		FeatureNamespace: cff.namespace,
		HelmChartName:    cff.HelmChartName,
		HelmChartDir:     filepath.Join(cff.HelmChartDir, cff.Dirs[componentName]),
	}
}

func NewTrafficManagementFeature(opts *FeatureOptions) *TrafficManagementFeature {
	cff := &CommonFeatureFields{
		FeatureOptions: *opts,
		enabled:        withOverrideBool(false, opts.InstallSpec.TrafficManagement.Enabled),
		namespace:      withOverrideString(opts.InstallSpec.DefaultNamespacePrefix, opts.InstallSpec.TrafficManagement.Components.Namespace),
	}
	cff.components = []component.Component{
		component.NewPilotComponent(newComponentOptions(cff, component.PilotComponentName)),
		component.NewSidecarComponent(newComponentOptions(cff, component.SidecarInjectorComponentName)),
	}

	return &TrafficManagementFeature{
		CommonFeatureFields: *cff,
	}
}

func (f *TrafficManagementFeature) RenderManifest() (string, util.Errors) {
	return renderComponents(f.components)
}

type SecurityFeature struct {
	CommonFeatureFields
}

func NewSecurityFeature(opts *FeatureOptions) *SecurityFeature {
	cff := &CommonFeatureFields{
		FeatureOptions: *opts,
		enabled:        withOverrideBool(false, opts.InstallSpec.Security.Enabled),
		namespace:      withOverrideString(opts.InstallSpec.DefaultNamespacePrefix, opts.InstallSpec.Security.Components.Namespace),
	}
	cff.components = []component.Component{
		component.NewCitadelComponent(newComponentOptions(cff, component.CitadelComponentName)),
		component.NewCertManagerComponent(newComponentOptions(cff, component.CertManagerComponentName)),
		component.NewNodeAgentComponent(newComponentOptions(cff, component.NodeAgentComponentName)),
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

func NewPolicyFeature(opts *FeatureOptions) *PolicyFeature {
	cff := &CommonFeatureFields{
		FeatureOptions: *opts,
		enabled:        withOverrideBool(false, opts.InstallSpec.Policy.Enabled),
		namespace:      withOverrideString(opts.InstallSpec.DefaultNamespacePrefix, opts.InstallSpec.Policy.Components.Namespace),
	}
	cff.components = []component.Component{
		component.NewPolicyComponent(newComponentOptions(cff, component.PolicyComponentName)),
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

func NewTelemetryFeature(opts *FeatureOptions) *TelemetryFeature {
	cff := &CommonFeatureFields{
		FeatureOptions: *opts,
		enabled:        withOverrideBool(false, opts.InstallSpec.Telemetry.Enabled),
		namespace:      withOverrideString(opts.InstallSpec.DefaultNamespacePrefix, opts.InstallSpec.Telemetry.Components.Namespace),
	}
	cff.components = []component.Component{
		component.NewTelemetryComponent(newComponentOptions(cff, component.TelemetryComponentName)),
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

func NewConfigManagementFeature(opts *FeatureOptions) *ConfigManagementFeature {
	cff := &CommonFeatureFields{
		FeatureOptions: *opts,
		enabled:        withOverrideBool(false, opts.InstallSpec.Telemetry.Enabled),
		namespace:      withOverrideString(opts.InstallSpec.DefaultNamespacePrefix, opts.InstallSpec.Telemetry.Components.Namespace),
	}
	cff.components = []component.Component{
		component.NewGalleyComponent(newComponentOptions(cff, component.GalleyComponentName)),
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

func NewAutoInjectionFeature(opts *FeatureOptions) *AutoInjectionFeature {
	cff := &CommonFeatureFields{
		FeatureOptions: *opts,
		enabled:        withOverrideBool(false, opts.InstallSpec.Telemetry.Enabled),
		namespace:      withOverrideString(opts.InstallSpec.DefaultNamespacePrefix, opts.InstallSpec.Telemetry.Components.Namespace),
	}
	cff.components = []component.Component{
		component.NewSidecarInjectorComponent(newComponentOptions(cff, component.SidecarInjectorComponentName)),
	}
	return &AutoInjectionFeature{
		CommonFeatureFields: *cff,
	}
}

func (f *AutoInjectionFeature) RenderManifest() (string, util.Errors) {
	return renderComponents(f.components)
}

func renderComponents(cs []component.Component) (manifest string, errsOut util.Errors) {
	var sb strings.Builder
	for _, c := range cs {
		s, errs := c.RenderManifest()
		errsOut = util.AppendErrs(errsOut, errs)
		_, err := sb.WriteString(s)
		errsOut = util.AppendErr(errsOut, err)
	}
	if len(errsOut) > 0 {
		return "", errsOut
	}
	return sb.String(), nil
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
