package feature

import (
	"github.com/ostromart/istio-installer/pkg/apis/istio/v1alpha2"
	"github.com/ostromart/istio-installer/pkg/component/component"
	"github.com/ostromart/istio-installer/pkg/name"
	"github.com/ostromart/istio-installer/pkg/translate"
	"github.com/ostromart/istio-installer/pkg/util"
)

// IstioFeature is a feature corresponding to Istio features defined in the IstioControlPlane proto.
type IstioFeature interface {
	// Run starts the Istio feature operation. Must be called before feature can be used.
	Run() error
	// RenderManifest returns a manifest string rendered against the IstioControlPlane parameters.
	RenderManifest() (name.ManifestMap, util.Errors)
}

// FeatureOptions are options for IstioFeature.
type FeatureOptions struct {
	// InstallSpec is the installation spec for the control plane.
	InstallSpec *v1alpha2.IstioControlPlaneSpec
	// Translator is the translator for this feature.
	Traslator *translate.Translator
}

// CommonFeatureFields
type CommonFeatureFields struct {
	// FeatureOptions is an embedded struct.
	FeatureOptions
	// components is a slice of components that are part of the feature.
	components []component.IstioComponent
}

// BaseFeature is the base feature, containing essential Istio base items.
type BaseFeature struct {
	// CommonFeatureFields is the struct shared among all features.
	CommonFeatureFields
}

// NewBaseFeature creates a new BaseFeature and returns a pointer to it.
func NewBaseFeature(opts *FeatureOptions) *BaseFeature {
	cff := &CommonFeatureFields{
		FeatureOptions: *opts,
	}
	cff.components = []component.IstioComponent{
		component.NewCRDComponent(newComponentOptions(cff, name.IstioBaseFeatureName)),
	}

	return &BaseFeature{
		CommonFeatureFields: *cff,
	}
}

// Run implements the IstioFeature interface.
func (f *BaseFeature) Run() error {
	return runComponents(f.components)
}

// RenderManifest implements the IstioFeature interface.
func (f *BaseFeature) RenderManifest() (name.ManifestMap, util.Errors) {
	return renderComponents(f.components)
}

// TrafficManagementFeature is the traffic management feature.
type TrafficManagementFeature struct {
	// CommonFeatureFields is the struct shared among all features.
	CommonFeatureFields
}

// NewTrafficManagementFeature creates a new TrafficManagementFeature and returns a pointer to it.
func NewTrafficManagementFeature(opts *FeatureOptions) *TrafficManagementFeature {
	cff := &CommonFeatureFields{
		FeatureOptions: *opts,
	}
	cff.components = []component.IstioComponent{
		component.NewPilotComponent(newComponentOptions(cff, name.TrafficManagementFeatureName)),
		component.NewSidecarInjectorComponent(newComponentOptions(cff, name.TrafficManagementFeatureName)),
	}

	return &TrafficManagementFeature{
		CommonFeatureFields: *cff,
	}
}

// Run implements the IstioFeature interface.
func (f *TrafficManagementFeature) Run() error {
	return runComponents(f.components)
}

// RenderManifest implements the IstioFeature interface.
func (f *TrafficManagementFeature) RenderManifest() (name.ManifestMap, util.Errors) {
	return renderComponents(f.components)
}

// SecurityFeature is the security feature.
type SecurityFeature struct {
	CommonFeatureFields
}

// NewSecurityFeature creates a new SecurityFeature and returns a pointer to it.
func NewSecurityFeature(opts *FeatureOptions) *SecurityFeature {
	cff := &CommonFeatureFields{
		FeatureOptions: *opts,
	}
	cff.components = []component.IstioComponent{
		component.NewCitadelComponent(newComponentOptions(cff, name.SecurityFeatureName)),
		component.NewCertManagerComponent(newComponentOptions(cff, name.SecurityFeatureName)),
		component.NewNodeAgentComponent(newComponentOptions(cff, name.SecurityFeatureName)),
	}
	return &SecurityFeature{
		CommonFeatureFields: *cff,
	}
}

// Run implements the IstioFeature interface.
func (f *SecurityFeature) Run() error {
	return runComponents(f.components)
}

// RenderManifest implements the IstioFeature interface.
func (f *SecurityFeature) RenderManifest() (name.ManifestMap, util.Errors) {
	return renderComponents(f.components)
}

// PolicyFeature is the policy feature.
type PolicyFeature struct {
	CommonFeatureFields
}

// NewPolicyFeature creates a new PolicyFeature and returns a pointer to it.
func NewPolicyFeature(opts *FeatureOptions) *PolicyFeature {
	cff := &CommonFeatureFields{
		FeatureOptions: *opts,
	}
	cff.components = []component.IstioComponent{
		component.NewPolicyComponent(newComponentOptions(cff, name.PolicyFeatureName)),
	}
	return &PolicyFeature{
		CommonFeatureFields: *cff,
	}
}

// Run implements the IstioFeature interface.
func (f *PolicyFeature) Run() error {
	return runComponents(f.components)
}

// RenderManifest implements the IstioFeature interface.
func (f *PolicyFeature) RenderManifest() (name.ManifestMap, util.Errors) {
	return renderComponents(f.components)
}

// TelemetryFeature is the telemetry feature.
type TelemetryFeature struct {
	CommonFeatureFields
}

// Run implements the IstioFeature interface.
func (f *TelemetryFeature) Run() error {
	return runComponents(f.components)
}

// NewTelemetryFeature creates a new TelemetryFeature and returns a pointer to it.
func NewTelemetryFeature(opts *FeatureOptions) *TelemetryFeature {
	cff := &CommonFeatureFields{
		FeatureOptions: *opts,
	}
	cff.components = []component.IstioComponent{
		component.NewTelemetryComponent(newComponentOptions(cff, name.TelemetryFeatureName)),
	}
	return &TelemetryFeature{
		CommonFeatureFields: *cff,
	}
}

// RenderManifest implements the IstioFeature interface.
func (f *TelemetryFeature) RenderManifest() (name.ManifestMap, util.Errors) {
	return renderComponents(f.components)
}

// ConfigManagementFeature is the config management feature.
type ConfigManagementFeature struct {
	CommonFeatureFields
}

// NewConfigManagementFeature creates a new ConfigManagementFeature and returns a pointer to it.
func NewConfigManagementFeature(opts *FeatureOptions) *ConfigManagementFeature {
	cff := &CommonFeatureFields{
		FeatureOptions: *opts,
	}
	cff.components = []component.IstioComponent{
		component.NewGalleyComponent(newComponentOptions(cff, name.ConfigManagementFeatureName)),
	}
	return &ConfigManagementFeature{
		CommonFeatureFields: *cff,
	}
}

// Run implements the IstioFeature interface.
func (f *ConfigManagementFeature) Run() error {
	return runComponents(f.components)
}

// RenderManifest implements the IstioFeature interface.
func (f *ConfigManagementFeature) RenderManifest() (name.ManifestMap, util.Errors) {
	return renderComponents(f.components)
}

// AutoInjectionFeature is the auto injection feature.
type AutoInjectionFeature struct {
	CommonFeatureFields
}

// NewAutoInjectionFeature creates a new AutoInjectionFeature and returns a pointer to it.
func NewAutoInjectionFeature(opts *FeatureOptions) *AutoInjectionFeature {
	cff := &CommonFeatureFields{
		FeatureOptions: *opts,
	}
	cff.components = []component.IstioComponent{
		component.NewSidecarInjectorComponent(newComponentOptions(cff, name.AutoInjectionFeatureName)),
	}
	return &AutoInjectionFeature{
		CommonFeatureFields: *cff,
	}
}

// Run implements the IstioFeature interface.
func (f *AutoInjectionFeature) Run() error {
	return runComponents(f.components)
}

// RenderManifest implements the IstioFeature interface.
func (f *AutoInjectionFeature) RenderManifest() (name.ManifestMap, util.Errors) {
	return renderComponents(f.components)
}

// newComponentOptions creates a component.ComponentOptions ptr from the given parameters.
func newComponentOptions(cff *CommonFeatureFields, featureName name.FeatureName) *component.ComponentOptions {
	return &component.ComponentOptions{
		InstallSpec: cff.InstallSpec,
		FeatureName: featureName,
		Translator:  cff.Traslator,
	}
}

// runComponents calls Run on all components in a feature.
func runComponents(cs []component.IstioComponent) error {
	for _, c := range cs {
		if err := c.Run(); err != nil {
			return err
		}
	}
	return nil
}

// renderComponents calls render manifest for all components in a feature and concatenates the outputs.
func renderComponents(cs []component.IstioComponent) (manifests name.ManifestMap, errsOut util.Errors) {
	manifests = make(name.ManifestMap)
	for _, c := range cs {
		m, err := c.RenderManifest()
		errsOut = util.AppendErr(errsOut, err)
		manifests[c.Name()] = m
	}
	if len(errsOut) > 0 {
		return nil, errsOut
	}
	return
}
