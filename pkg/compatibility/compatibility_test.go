package compatibility

import (
	"fmt"
	"testing"

//"github.com/kr/pretty/"
	"github.com/kylelemons/godebug/diff"
	"github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
	"github.com/ostromart/istio-installer/pkg/util"
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

func TestProtoToValues(t *testing.T) {
	tests := []struct {
		desc    string
		proto   *v1alpha1.InstallerSpec
		want    string
		wantErr string
	}{
		{
			desc: "nil success",
			want: "",
		},
		{
			desc: "single leaf",
			proto: &v1alpha1.InstallerSpec{
				TrafficManagement: &v1alpha1.TrafficManagementConfig{
					ProxyConfig: &v1alpha1.ProxyConfig{
						StatusPort: toUint32Ptr(123),
					},
				},
			},
			want: `global:
  monitoringPort: 123
`,
		},
	}

	for _, tt := range tests {
		fmt.Println(tt.desc)
		t.Run(tt.desc, func(t *testing.T) {
			got, err := ProtoToValues(defaultMappings, tt.proto)
			if gotErr, wantErr := errToString(err), tt.wantErr; gotErr != wantErr {
				t.Errorf("ProtoToValues(%s)(%v): gotErr:%s, wantErr:%s", tt.desc, tt.proto, gotErr, wantErr)
			}
			if want := tt.want; err == nil && got != want {
				t.Errorf("ProtoToValues(%s) got:\n%s\nwant:\n%s\n", tt.desc, got, want)
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

// to ptr conversion utility functions
func toStringPtr(s string) *string { return &s }
func toUint32Ptr(i uint32) *uint32 { return &i }
