package validate

import (
	"bytes"
	"fmt"
	"github.com/ostromart/istio-installer/pkg/util"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/jsonpb"
	"github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		desc     string
		yamlStr  string
		wantErrs util.Errors
	}{
		{
			desc: "nil success",
		},
		{
			desc: "TrafficManagement",
			yamlStr: `
trafficManagement:
  enabled: true
  namespace: istio-system-traffic
  clusterDomain: "my.domain"
`,
		},
		{
			desc: "PilotConfig",
			yamlStr: `
trafficManagement:
  pilot:
    sidecar: true
`,
		},
		{
			desc: "SidecarInjectorConfig",
			yamlStr: `
trafficManagement:
  sidecarInjector:
    enableNamespacesByDefault: true
`,
		},
		{
			desc: "CommonConfig",
			yamlStr: `
hub: docker.io/istio
tag: v1.2.3
trafficManagement:
  proxy:
    common:
      enabled: true
      namespace: istio-control-system
      debug: INFO
      env:
        aa: bb
        cc: dd
      args:
        b: b
        c: d
      k8s:
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
        readinessProbe:
          initialDelaySeconds: 11
          periodSeconds: 22
          successThreshold: 33
          failureThreshold: 44
        hpaSpec:
          scaleTargetRef:
            apiVersion: apps/v1
            kind: Deployment
            name: php-apache
          minReplicas: 1
          maxReplicas: 10
          targetCPUUtilizationPercentage: 80
        nodeSelector:
          disktype: ssd
`,
		},
		{
			desc: "BadTag",
			yamlStr: `
hub: ?illegal-tag!
`,
			wantErrs: makeErrors([]string{`invalid value Hub:?illegal-tag!`}),
		},
		{
			desc: "BadHub",
			yamlStr: `
hub: docker.io:tag/istio
`,
			wantErrs: makeErrors([]string{`invalid value Hub:docker.io:tag/istio`}),
		},
		{
			desc: "GoodURL",
			yamlStr: `
installPackagePath: file://local/file/path
`,
		},
		{
			desc: "BadURL",
			yamlStr: `
installPackagePath: bad_schema://local/file/path
`,
			wantErrs: makeErrors([]string{`invalid value InstallPackagePath:bad_schema://local/file/path`}),
		},
	}

	for _, tt := range tests {
		fmt.Println(tt.desc)
		t.Run(tt.desc, func(t *testing.T) {
			ispec := &v1alpha1.InstallerSpec{}
			err := unmarshalWithJSONPB(tt.yamlStr, ispec)
			if err != nil {
				t.Fatalf("unmarshalWithJSONPB(%s): got error %s", tt.desc, err)
			}
			errs := ValidateInstallerSpec(defaultValidations, ispec)
			if gotErrs, wantErrs := errs, tt.wantErrs; !util.EqualErrors(gotErrs, wantErrs) {
				t.Errorf("ProtoToValues(%s)(%v): gotErrs:%s, wantErrs:%s", tt.desc, tt.yamlStr, gotErrs, wantErrs)
			}
		})
	}
}

func unmarshalWithJSONPB(y string, out proto.Message) error {
	jb, err := yaml.YAMLToJSON([]byte(y))
	if err != nil {
		return err
	}

	u := jsonpb.Unmarshaler{}
	err = u.Unmarshal(bytes.NewReader(jb), out)
	if err != nil {
		return err
	}
	return nil
}

// errToString returns the string representation of err and the empty string if
// err is nil.
func errToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
