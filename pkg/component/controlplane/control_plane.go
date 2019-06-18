package controlplane

import (
	"fmt"
	"strings"

	"github.com/ostromart/istio-installer/pkg/translate"

	"github.com/ostromart/istio-installer/pkg/component/feature"
	"github.com/ostromart/istio-installer/pkg/util"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	installerv1alpha1 "github.com/ostromart/istio-installer/pkg/apis/istio/v1alpha2"
)

// IstioControlPlane is an installation of an Istio control plane.
type IstioControlPlane struct {
	features []feature.IstioFeature
	started  bool
}

// NewIstioControlPlane creates a new IstioControlPlane and returns a pointer to it.
func NewIstioControlPlane(installSpec *installerv1alpha1.IstioControlPlaneSpec, translator *translate.Translator) *IstioControlPlane {
	opts := &feature.FeatureOptions{
		InstallSpec: installSpec,
		Traslator:   translator,
	}
	return &IstioControlPlane{
		features: []feature.IstioFeature{
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
func (i *IstioControlPlane) RenderManifest() (manifest string, errsOut util.Errors) {
	if !i.started {
		return "", util.NewErrs(fmt.Errorf("IstioControlPlane must be Run before calling RenderManifest"))
	}
	var sb strings.Builder
	for _, f := range i.features {
		s, errs := f.RenderManifest()
		errsOut = util.AppendErrs(errsOut, errs)
		_, err := sb.WriteString(s)
		errsOut = util.AppendErr(errsOut, err)
	}
	if len(errsOut) > 0 {
		return "", errsOut
	}
	return sb.String(), nil
}
