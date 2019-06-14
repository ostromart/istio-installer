package translate

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
	"github.com/ostromart/istio-installer/pkg/component/component"
	"github.com/ostromart/istio-installer/pkg/util"
)

func TestValuesOverlayToValues(t *testing.T) {
	tests := []struct {
		desc          string
		componentName string
		inYaml        string
		inStruct      interface{}
		want          string
		wantErr       string
	}{
		{
			desc: "pilot",
			inYaml: `
pilot:
  enabled: true
  resources:
    requests:
      cpu: 111m
      memory: 222Mi
`,
			inStruct: &v1alpha1.Values{},
			want:     ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if err := yaml.Unmarshal([]byte(tt.inYaml), tt.inStruct); err != nil {
				t.Fatalf("Unmarshal: %s", err)
			}
			got, err := ValuesOverlayToValues(component.PilotComponentName, tt.inStruct)
			if gotErr, wantErr := errToString(err), tt.wantErr; gotErr != wantErr {
				t.Fatalf("ValuesOverlayToValues(%s): gotErr:%s, wantErr:%s", tt.desc, gotErr, wantErr)
			}
			if want := tt.want; !util.IsYAMLEqual(got, want) {
				t.Errorf("ValuesOverlayToValues(%s): got:\n%s\n\nwant:\n%s\nDiff:\n%s\n", tt.desc, got, want, util.YAMLDiff(got, want))
			}
		})
	}
}

// errToString returns the string representation of err and the empty string if
// err is nil.
func errToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
