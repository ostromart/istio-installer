// Copyright 2019 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package translate

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ostromart/istio-installer/pkg/apis/istio/v1alpha2"
	"github.com/ostromart/istio-installer/pkg/version"

	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/jsonpb"
	"github.com/kr/pretty"
	"github.com/ostromart/istio-installer/pkg/util"
)

func TestProtoToValuesV12(t *testing.T) {
	tests := []struct {
		desc    string
		yamlStr string
		want    string
		wantErr string
	}{
		{
			desc: "nil success",
			want: `
certmanager:
  enabled: false
global:
  istioNamespace: ""
  policyNamespace: ""
  telemetryNamespace: ""
mixer:
  policy:
    enabled: false
  telemetry:
    enabled: false
nodeagent:
  enabled: false
pilot:
  enabled: false
security:
  enabled: false
`,
		},
		{
			desc: "global",
			yamlStr: `
hub: docker.io/istio
tag: 1.2.3
defaultNamespacePrefix: istio-system
security:
  components:
    namespace: istio-security
`,
			want: `
certmanager:
  enabled: false
global:
  hub: docker.io/istio
  istioNamespace: istio-system
  policyNamespace: istio-system
  tag: 1.2.3
  telemetryNamespace: istio-system
mixer:
  policy:
    enabled: false
  telemetry:
    enabled: false
nodeagent:
  enabled: false
pilot:
  enabled: false
security:
  enabled: false

`,
		},
		{
			desc: "Security",
			yamlStr: `
security:
  controlPlaneMtls: true
  dataPlaneMtlsStrict: false
`,
			want: `
global:
  controlPlaneSecurityEnabled: true
  mtls:
    enabled: false
`,
		},
		{
			desc: "SidecarInjector",
			yamlStr: `
trafficManagement:
  sidecarInjector:
    enableNamespacesByDefault: false
`,
			want: `
sidecarInjectorWebhook:
  enableNamespacesByDefault: false
`,
		},
	}

	tr := Translators[version.MinorVersion{1, 2}]
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			ispec := &v1alpha2.IstioControlPlaneSpec{}
			err := unmarshalWithJSONPB(tt.yamlStr, ispec)
			if err != nil {
				t.Fatalf("unmarshalWithJSONPB(%s): got error %s", tt.desc, err)
			}
			dbgPrint("ispec: \n%s\n", pretty.Sprint(ispec))
			got, err := tr.ProtoToValues(ispec)
			if gotErr, wantErr := errToString(err), tt.wantErr; gotErr != wantErr {
				t.Errorf("ProtoToValues(%s)(%v): gotErr:%s, wantErr:%s", tt.desc, tt.yamlStr, gotErr, wantErr)
			}
			if want := tt.want; !util.IsYAMLEqual(got, want) {
				t.Errorf("ProtoToValues(%s): got:\n%s\n\nwant:\n%s\nDiff:\n%s\n", tt.desc, got, want, util.YAMLDiff(got, want))
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

func stripNL(s string) string {
	return strings.Trim(s, "\n")
}
