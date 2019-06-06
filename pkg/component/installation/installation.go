package installation

import (
	"strings"

	"github.com/ostromart/istio-installer/pkg/component/component"
	"github.com/ostromart/istio-installer/pkg/component/feature"
	"github.com/ostromart/istio-installer/pkg/util"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	installerv1alpha1 "github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
)

type Installation interface {
	Run() error
	RenderManifest() (string, util.Errors)
}
type InstallationImpl struct {
	features []feature.Feature
}

func NewInstallation(installSpec *installerv1alpha1.IstioControlPlaneSpec, helmChartName, helmChartDir, globalValuesFilePath string, dirs component.ComponentDirLayout) Installation {
	opts := &feature.FeatureOptions{
		InstallSpec:      installSpec,
		HelmChartName:    helmChartName,
		HelmChartDir:     helmChartDir,
		GlobalValuesFile: globalValuesFilePath,
		Dirs:             dirs,
	}
	return &InstallationImpl{
		features: []feature.Feature{
			feature.NewTrafficManagementFeature(opts),
			feature.NewSecurityFeature(opts),
			feature.NewPolicyFeature(opts),
			feature.NewTelemetryFeature(opts),
			feature.NewConfigManagementFeature(opts),
			feature.NewAutoInjectionFeature(opts),
		},
	}
}

func (i InstallationImpl) Run() error {
	for _, f := range i.features {
		if err := f.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (i InstallationImpl) RenderManifest() (manifest string, errsOut util.Errors) {
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
