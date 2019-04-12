package translate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/jsonpb"
	"github.com/kylelemons/godebug/diff"
	"github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
	"github.com/ostromart/istio-installer/pkg/util"
	k8sjson "k8s.io/apimachinery/pkg/util/json"
	//	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	//	k8syaml "k8s.io/client-go/pkg/util/yaml"
	"github.com/kr/pretty"
)

func TestSetYAML(t *testing.T) {
	tests := []struct {
		desc    string
		root    util.Tree
		path    string
		value   string
		want    string
		wantErr string
	}{
		{
			desc:    "insert no path",
			path:    "",
			value:   "val1",
			want:    `val1`,
			wantErr: "path cannot be empty",
		},
		{
			desc:  "insert empty",
			path:  "a/b/c",
			value: "val1",
			want: `a:
  b:
    c: val1
`,
		},
		{
			desc: "overwrite",
			root: util.Tree{
				"a": util.Tree{
					"b": util.Tree{
						"c": "val1",
					},
				},
			},
			path:  "a/b/c",
			value: "val2",
			want: `a:
  b:
    c: val2
`,
		},
		{
			desc: "partial create",
			root: util.Tree{
				"a": util.Tree{
					"b": util.Tree{
						"c": "val1",
					},
				},
			},
			path:  "a/b/d",
			value: "val2",
			want: `a:
  b:
    c: val1
    d: val2
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			root := tt.root
			if root == nil {
				root = make(util.Tree)
			}
			p := util.PathFromString(tt.path)
			err := setYAML(root, p, tt.value)
			fmt.Println(err)
			if gotErr, wantErr := errToString(err), tt.wantErr; gotErr != wantErr {
				t.Errorf("TestSetYAML()%s: gotErr:%s, wantErr:%s", tt.desc, gotErr, wantErr)
				return
			}
			if got, want := root.String(), tt.want; err == nil && got != want {
				t.Errorf("TestSetYAML(%s) got:\n%s\nwant:\n%s\ndiff:\n%s\n", tt.desc, got, want, diff.Diff(got, want))
			}
		})
	}
}

func TestUnmarshalKubernetes(t *testing.T) {
	tests := []struct {
		desc    string
		yamlStr string
		want    string
	}{
		{
			desc:    "nil success",
			yamlStr: "",
			want:    "{}",
		},
		{
			desc: "hpaSpec",
			yamlStr: `
hpaSpec:
  maxReplicas: 10
  minReplicas: 1
  targetCPUUtilizationPercentage: 80
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: php-apache
`,
		},
		{
			desc: "resources",
			yamlStr: `
resources:
  limits:
    cpu: 444m
    memory: 333Mi
  requests:
    cpu: 222m
    memory: 111Mi
`,
		},
		{
			desc: "podDisruptionBudget",
			yamlStr: `
podDisruptionBudget:
  maxUnavailable: 1
  selector:
    matchLabels:
      app: pilot
`,
		},
		{
			desc: "readinessProbe",
			yamlStr: `
readinessProbe:
  failureThreshold: 44
  initialDelaySeconds: 11
  periodSeconds: 22
  successThreshold: 33
  handler: {}
`,
		},
		{
			desc: "k8sObjectOverride",
			yamlStr: `
k8sObjectOverride:
- patchType: JSON
  op: PATCH
  apiVersion: v1
  kind: Service
  metadata:
    name: istio-pilot
    namespace: istio-system
  data:
    spec:
      ports:
      - port: 11111
        name: grpc-xds # direct
      - port: 22222
        name: https-xds # mTLS
      - port: 33333
        name: http-legacy-discovery # direct
      - port: 44444
        name: http-monitoring
`,
		},
	}
	for _, tt := range tests {
		fmt.Println(tt.desc)
		t.Run(tt.desc, func(t *testing.T) {
			tk := &v1alpha1.TestKube{}
			err := unmarshalWithJSONPB(tt.yamlStr, tk)
			if err != nil {
				t.Fatalf("unmarshalWithJSONPB(%s): got error %s", tt.desc, err)
			}
			s, err := marshalWithJSONPB(tk)
			if err != nil {
				t.Fatalf("unmarshalWithJSONPB(%s): got error %s", tt.desc, err)
			}
			got, want := stripNL(s), stripNL(tt.want)
			if want == "" {
				want = stripNL(tt.yamlStr)
			}
			if !isYAMLEqual(got, want) {
				t.Errorf("%s: got:\n%s\nwant:\n%s\n(-got, +want)\n%s\n", tt.desc, got, want, diff.Diff(got, want))
			}
		})
	}
}

func stripNL(s string) string {
	return strings.Trim(s, "\n")
}

func isYAMLEqual(a, b string) bool {
	ajb, err := yaml.YAMLToJSON([]byte(a))
	if err != nil {
		fmt.Printf("bad YAML in isYAMLEqual:\n%s\n", a)
		return false
	}
	bjb, err := yaml.YAMLToJSON([]byte(b))
	if err != nil {
		fmt.Printf("bad YAML in isYAMLEqual:\n%s\n", b)
		return false
	}

	//fmt.Printf("a:\n%s\nb:\n%s\n", string(ajb), string(bjb))
	return string(ajb) == string(bjb)
}

func TestProtoToValues(t *testing.T) {
	tests := []struct {
		desc string
		// mappings defaults to defaultMappings
		mappings map[string]*Translation
		yamlStr  string
		want     string
		wantErr  string
	}{
		{
			desc: "nil success",
			want: "",
		},
		{
			desc: "PilotConfig",
			yamlStr: `
trafficManagement:
  pilot:
    sidecar: true
`,
			want: `
global:
  pilot:
    sidecar: true
`,
		},
		{
			desc: "ProxyConfig",
			yamlStr: `
trafficManagement:
  includeIpRanges: "1.1.0.0/16,2.2.0.0/16"
  excludeIpRanges: "3.3.0.0/16,4.4.0.0/16"
  includeInboundPorts: "111,222"
  excludeInboundPorts: "333,444"
  clusterDomain: "my.domain"
  podDnsSearchNamespaces: "my-namespace"
  enableAutoInjection: true
  enableNamespacesByDefault: true
  proxy:
    interceptionMode: TPROXY
    connectTimeout: "11s"
    drainDuration: "22s"
    parentShutdownDuration : "33s"
    concurrency: 5
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
			want: `
global:
  proxy: 
    additionalArgs:
      b: b
      c: d
    clusterDomain: my.domain
    concurrency: 5
    debug: true
    drainDuration: 22s
    enableCoredump: true
    env:
      aa: bb
      cc: dd
    excludeInboundPorts: 333,444
    excludeIpRanges: 3.3.0.0/16,4.4.0.0/16
    imagePullPolicy: IfNotPresent
    includeInboundPorts: 111,222
    includeIpRanges: 1.1.0.0/16,2.2.0.0/16
    interceptionMode: 1
    parentShutdownDuration: 33s
    podDnsSearchNamespaces: my-namespace
    privileged: true
    proxyInitImage: proxy_init_image
    readinessProbe:
      failureThreshold: 44
      initialDelaySeconds: 11
      periodSeconds: 22
      successThreshold: 33
      timeoutSeconds: 0
    resources:
      limits:
        cpu: 500m
        memory: 128Mi
      requests:
        cpu: 250m
        memory: 64Mi

`,
		},
	}

	for _, tt := range tests {
		fmt.Println(tt.desc)
		t.Run(tt.desc, func(t *testing.T) {
			mappings := tt.mappings
			if mappings == nil {
				mappings = defaultMappings
			}
			ispec := &v1alpha1.InstallerSpec{}
			err := unmarshalWithJSONPB(tt.yamlStr, ispec)
			if err != nil {
				t.Fatalf("unmarshalWithJSONPB(%s): got error %s", tt.desc, err)
			}
			fmt.Println("ispec: ", pretty.Sprint(ispec))
			got, err := ProtoToValues(mappings, ispec)
			fmt.Println(got)
			if gotErr, wantErr := errToString(err), tt.wantErr; gotErr != wantErr {
				t.Errorf("ProtoToValues(%s)(%v): gotErr:%s, wantErr:%s", tt.desc, tt.yamlStr, gotErr, wantErr)
			}
			if got, want := stripNL(got), stripNL(tt.want); err == nil && !isYAMLEqual(got, want) {
				t.Errorf("ProtoToValues(%s) got:\n%s\nwant:\n%s\n", tt.desc, got, want)
			}
		})
	}
}

func marshalYAML(in interface{}) (string, error) {
	out, err := yaml.Marshal(in)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func marshalWithJSONPB(in *v1alpha1.TestKube) (string, error) {
	m := jsonpb.Marshaler{}
	js, err := m.MarshalToString(in)
	if err != nil {
		return "", err
	}
	yb, err := yaml.JSONToYAML([]byte(js))
	if err != nil {
		return "", err
	}
	return string(yb), nil
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

func unmarshalWithJSON(y string, out interface{}) error {
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

func unmarshalWithKubernetes(yaml string, out interface{}) error {
	r := bytes.NewReader([]byte(yaml))
	decoder := k8syaml.NewYAMLOrJSONDecoder(r, 1024)

	err := decoder.Decode(out)
	if err != nil {
		return fmt.Errorf("error decoding object: %s\n%s\n", err, yaml)
	}
	return nil
}
func marshalWithKubernetes(in interface{}) (string, error) {
	jb, err := k8sjson.Marshal(in)
	if err != nil {
		return "", err
	}
	yb, err := yaml.JSONToYAML(jb)
	if err != nil {
		return "", err
	}
	return string(yb), nil
}

func unmarshalWithKubernetesThroughJSON(y string, out interface{}) error {
	jb, err := yaml.YAMLToJSON([]byte(y))
	if err != nil {
		return err
	}

	decoder := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(jb), 1024)
	err = decoder.Decode(out)
	if err != nil {
		return fmt.Errorf("error decoding object: %s\n%s\n", err, y)
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

// to ptr conversion utility functions
func toStringPtr(v string) *string { return &v }
func toBoolPtr(v bool) *bool       { return &v }
func toUint32Ptr(v uint32) *uint32 { return &v }
