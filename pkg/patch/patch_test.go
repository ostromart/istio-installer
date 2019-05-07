package patch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kr/pretty"
	"github.com/ostromart/istio-installer/pkg/util"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/jsonpb"
	"github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
)

/*
apiVersion: v1
kind: Deployment
name: istio-galley
namespace: istio-system
- op: replace
  path: spec/template/spec/containers[name: galley]/imagePullPolicy
  value: Always
- op: replace
  path: spec/template/spec/containers[name: galley]/ports[containerPort: 443]
  value: 123
- op: add
  path: spec/template/spec/containers[name: galley]/ports
  value: --newFlag=true
- op: add
  path: spec/template/spec/containers
  value:
    name: newContainer
    command:
    - /bin/bash
    - echo "Hello world"
*/

func TestGetNode(t *testing.T) {
	/*tr := map[string]interface {}{
		"a": map[string]interface {}{
			"b": map[string]interface {}{
				"c": []interface {}{
					map[string]interface {}{
						"d":  "1",
						"dd": "11",
					},
					map[string]interface {}{
						"e":  "2",
						"ee": "22",
					},
				},
			},
		},
	}*/

	tests := []struct {
		desc        string
		yAML        string
		path        string
		createNodes bool
		want        interface{}
		wantErr     string
	}{
		{
			desc: "simple path",
			yAML: `
a:
  b:
    c: d
`,
			path: "a.b.c",
			want: "d",
		},
		{
			desc: "leaf list",
			yAML: `a:
  b:
    c:
    - d: 1
      dd: 11
    - e: 2
      ee: 22
`,
			path: "a.b.c.d:1",
			want: "0",
		},
	}

	for _, tt := range tests {
		fmt.Println(tt.desc)
		t.Run(tt.desc, func(t *testing.T) {
			root := make(map[string]interface{})
			err := yaml.Unmarshal([]byte(tt.yAML), &root)
			fmt.Println(pretty.Sprint(root))
			if err != nil {
				t.Fatalf("yaml.Unmarshal(%s): got error %s", tt.desc, err)
			}
			inc, err := getNode(makeNodeContext(root), util.PathFromString(tt.path))
			if gotErr, wantErr := errToString(err), tt.wantErr; gotErr != wantErr {
				t.Fatalf("%s: gotErr:%s, wantErr:%s", tt.desc, gotErr, wantErr)
			}
			if got, want := fmt.Sprint(inc.node), tt.want; got != want {
				t.Errorf("%s: got:\n%s\n\nwant:\n%s", tt.desc, got, want)
			}
		})
	}
}

func TestPatchYAMLManifest(t *testing.T) {
	tests := []struct {
		desc    string
		base    string
		overlay string
		want    string
		wantErr string
	}{
		{
			desc: "nil success",
		},
		{
			desc: "Deployment",
			base: `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-citadel
  namespace: istio-system
spec:
  template:
    spec:
      containers:
      - name: deleteThis
        foo: bar
      - name: galley
        ports:
        - containerPort: 443
        - containerPort: 15014
        - containerPort: 9901
        command:
        - /usr/local/bin/galley
        - server
        - --meshConfigFile=/etc/mesh-config/mesh
        - --livenessProbeInterval=1s
        - --validation-webhook-config-file
`,
			overlay: `
overlays:
- kind: Deployment
  name: istio-citadel
  patches:
    - path: spec.template.spec.containers.name:deleteThis
`,
			want: `
`,
		},
	}
/*
    - path: spec.template.spec.containers.name:galley.command.:--validation-webhook-config-file
    - path: spec.template.spec.containers.name:galley.command.:--livenessProbeInterval=1s
      value: --livenessProbeInterval=1111s
    - path: spec.template.spec.containers.name:galley.ports.containerPort:15014
      value: 22222
    - path: spec.template.spec.containers.name:galley.ports.containerPort:9901
    - path: spec.template.spec.containers.name:deleteThis

*/

	for _, tt := range tests {
		fmt.Println(tt.desc)
		t.Run(tt.desc, func(t *testing.T) {
			rc := &v1alpha1.KubernetesResourcesSpec{}
			err := unmarshalWithJSONPB(tt.overlay, rc)
			if err != nil {
				t.Fatalf("unmarshalWithJSONPB(%s): got error %s", tt.desc, err)
			}
			//fmt.Println(pretty.Sprint(rc))
			got, err := PatchYAMLManifest(tt.base, "istio-system", rc.Overlays)
			if gotErr, wantErr := errToString(err), tt.wantErr; gotErr != wantErr {
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
