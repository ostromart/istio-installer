package controlplane

import (
	"fmt"

	"github.com/ostromart/istio-installer/pkg/apis/istio/v1alpha2"
	"github.com/ostromart/istio-installer/pkg/component/feature"
	"github.com/ostromart/istio-installer/pkg/name"
	"github.com/ostromart/istio-installer/pkg/translate"
	"github.com/ostromart/istio-installer/pkg/util"
)

// IstioControlPlane is an installation of an Istio control plane.
type IstioControlPlane struct {
	features []feature.IstioFeature
	started  bool
}

// NewIstioControlPlane creates a new IstioControlPlane and returns a pointer to it.
func NewIstioControlPlane(installSpec *v1alpha2.IstioControlPlaneSpec, translator *translate.Translator) *IstioControlPlane {
	opts := &feature.FeatureOptions{
		InstallSpec: installSpec,
		Traslator:   translator,
	}
	return &IstioControlPlane{
		features: []feature.IstioFeature{
			feature.NewBaseFeature(opts),
			feature.NewTrafficManagementFeature(opts),
			feature.NewSecurityFeature(opts),
			feature.NewPolicyFeature(opts),
			feature.NewTelemetryFeature(opts),
			feature.NewConfigManagementFeature(opts),
			feature.NewAutoInjectionFeature(opts),
		},
	}
}

// Run starts the Istio control plane.
func (i *IstioControlPlane) Run() error {
	for _, f := range i.features {
		if err := f.Run(); err != nil {
			return err
		}
	}
	i.started = true
	return nil
}

// RenderManifest returns a manifest rendered against
func (i *IstioControlPlane) RenderManifest() (manifests name.ManifestMap, errsOut util.Errors) {
	if !i.started {
		return nil, util.NewErrs(fmt.Errorf("IstioControlPlane must be Run before calling RenderManifest"))
	}

	manifests = make(name.ManifestMap)
	for _, f := range i.features {
		ms, errs := f.RenderManifest()
		manifests = mergeManifestMaps(manifests, ms)
		errsOut = util.AppendErrs(errsOut, errs)
	}
	if len(errsOut) > 0 {
		return nil, errsOut
	}
	return
}

func mergeManifestMaps(a, b name.ManifestMap) name.ManifestMap {
	out := make(name.ManifestMap)
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}
