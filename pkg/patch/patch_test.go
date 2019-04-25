package patch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ostromart/istio-installer/pkg/util"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/jsonpb"
	"github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
)

func TestPatchYAMLManifest(t *testing.T) {
	tests := []struct {
		desc     string
		base     string
		overlay  string
		want     string
		wantErr  string
	}{
		{
			desc: "nil success",
		},
		{
			desc: "Service",
			base: `
# Source: istio/charts/prometheus/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: promsd
  namespace: istio-system
  annotations:
    prometheus.io/scrape: 'true'
  labels:
    kubernetes.io/cluster-service: "true"
    name: promsd
`,
			overlay: `
overlays:
- patchType: JSON
  op: MERGE
  data: 
    apiVersion: v1
    kind: Service
    metadata:
      name: promsd
      namespace: istio-system
      annotations:
        prometheus.io/scrape: 'false'
`,
			want: `
apiVersion: v1
kind: Service
metadata:
  name: promsd
  namespace: istio-system
  annotations:
    prometheus.io/scrape: 'false'
  labels:
    kubernetes.io/cluster-service: "true"
    name: promsd
`,
		},
	}

	for _, tt := range tests {
		fmt.Println(tt.desc)
		t.Run(tt.desc, func(t *testing.T) {
			rc := &v1alpha1.KubernetesResourcesConfig{}
			err := unmarshalWithJSONPB(tt.overlay, rc)
			if err != nil {
				t.Fatalf("unmarshalWithJSONPB(%s): got error %s", tt.desc, err)
			}
			got, err := PatchYAMLManifest(tt.base, rc.Overlays)
			if gotErr, wantErr := errToString(err), tt.wantErr; gotErr != wantErr  {
				t.Fatalf("PatchYAMLManifest(%s): gotErr:%s, wantErr:%s", tt.desc, gotErr, wantErr)
			}
			if want := tt.want; !util.IsYAMLEqual(got, want) {
				t.Errorf("PatchYAMLManifest(%s): got:\n%s\n\nwant:\n%s", tt.desc, got, want)
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

func unmarshalWithJSON(y string, out proto.Message) error {
	jb, err := yaml.YAMLToJSON([]byte(y))
	if err != nil {
		return err
	}

	err = json.Unmarshal(jb, out)
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
