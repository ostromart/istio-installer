package installation

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ostromart/istio-installer/pkg/manifest"

	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/jsonpb"
	"github.com/kylelemons/godebug/diff"
	"github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
	"github.com/ostromart/istio-installer/pkg/component/component"
)

var (
	testDataDir      string
	helmChartTestDir string
	globalValuesFile string
)

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	testDataDir = filepath.Join(wd, "testdata")
	helmChartTestDir = filepath.Join(testDataDir, "charts")
	globalValuesFile = filepath.Join(helmChartTestDir, "global.yaml")
}

func TestRenderInstallationSuccess(t *testing.T) {
	tests := []struct {
		desc        string
		installSpec string
		wantFile    string
	}{
		{
			desc: "all_off",
			installSpec: `
customPackagePath: file://foo
trafficManagement:
  enabled: false
policy:
  enabled: false
telemetry:
  enabled: false
security:
  enabled: false
configManagement:
  enabled: false
autoInjection:
  enabled: false
`,
		},
		{
			desc: "pilot_default",
			installSpec: `
customPackagePath: file://foo
trafficManagement:
  enabled: true
  components:
    proxy:
      common:
        enabled: false
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			var is v1alpha1.IstioControlPlaneSpec
			err := unmarshalWithJSONPB(tt.installSpec, &is)
			if err != nil {
				t.Fatalf("yaml.Unmarshal(%s): got error %s", tt.desc, err)
			}

			ins := NewInstallation(&is, "istio", helmChartTestDir, globalValuesFile, component.V12DirLayout)
			if err = ins.Run(); err != nil {
				t.Fatal(err)
			}

			got, errs := ins.RenderManifest()
			if len(errs) != 0 {
				t.Fatal(errs.Error())
			}
			want, err := readFile(tt.desc + ".yaml")
			if err != nil {
				t.Fatal(err)
			}
			//fmt.Println(got)
			diff, err := ManifestDiff(got, want)
			if err != nil {
				t.Fatal(err)
			}
			if diff != "" {

				//t.Errorf("got objects:\n%s\nwant objects:\n%s\n", ObjectsInManifest(got), ObjectsInManifest(want))
				t.Errorf("%s: got:\n%s\nwant:\n%s\n(-got, +want)\n%s\n", tt.desc, "", "", diff)
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

func readFile(path string) (string, error) {
	b, err := ioutil.ReadFile(filepath.Join(testDataDir, path))
	return string(b), err
}

func stripNL(s string) string {
	return strings.Trim(s, "\n")
}

// TODO: move to util
func IsYAMLEqual(a, b string) bool {
	if strings.TrimSpace(a) == "" && strings.TrimSpace(b) == "" {
		return true
	}
	ajb, err := yaml.YAMLToJSON([]byte(a))
	if err != nil {
		return false
	}
	bjb, err := yaml.YAMLToJSON([]byte(b))
	if err != nil {
		return false
	}

	return string(ajb) == string(bjb)
}

func YAMLDiff(a, b string) string {
	ao, bo := make(map[string]interface{}), make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(a), &ao); err != nil {
		return err.Error()
	}
	if err := yaml.Unmarshal([]byte(b), &bo); err != nil {
		return err.Error()
	}

	ay, err := yaml.Marshal(ao)
	if err != nil {
		return err.Error()
	}
	by, err := yaml.Marshal(bo)
	if err != nil {
		return err.Error()
	}

	return diff.Diff(string(ay), string(by))
}

func ManifestDiff(a, b string) (string, error) {
	ao, err := manifest.ParseObjectsFromYAMLManifest(context.TODO(), a)
	if err != nil {
		return "", err
	}
	bo, err := manifest.ParseObjectsFromYAMLManifest(context.TODO(), b)
	if err != nil {
		return "", err
	}
	aom, bom := ao.ToMap(), bo.ToMap()
	var sb strings.Builder
	for ak, av := range aom {
		ay, err := av.YAML()
		if err != nil {
			return "", err
		}
		by, err := bom[ak].YAML()
		if err != nil {
			return "", err
		}
		diff := YAMLDiff(string(ay), string(by))
		if diff != "" {
			sb.WriteString("\n\nObject " + ak + " has diffs:\n\n")
			sb.WriteString(diff)
		}
	}
	for bk, bv := range bom {
		if aom[bk] == nil {
			by, err := bv.YAML()
			if err != nil {
				return "", err
			}
			diff := YAMLDiff(string(by), "")
			if diff != "" {
				sb.WriteString("\n\nObject " + bk + " is missing:\n\n")
				sb.WriteString(diff)
			}
		}
	}
	return sb.String(), err
}

func ObjectsInManifest(mstr string) string {
	ao, err := manifest.ParseObjectsFromYAMLManifest(context.TODO(), mstr)
	if err != nil {
		return err.Error()
	}
	var out []string
	for _, v := range ao.Items {
		out = append(out, v.Hash())
	}
	return strings.Join(out, "\n")
}
