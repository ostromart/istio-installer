// Package compatibility defines translations from installer proto to values.yaml.
package compatibility

import (
	"path/filepath"
	"reflect"
	"strings"

	"github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
	"fmt"
	"gopkg.in/yaml.v2"
)

// Tree is a tree.
type Tree map[string]interface{}

// TranslationFunc maps a proto API path into a YAML values tree.
type TranslationFunc func(t *Translation, root Tree, value string) error

// Translation is a mapping between a proto data path and a YAML values path.
type Translation struct {
	yAMLPath        string
	translationFunc TranslationFunc
}

var (
	mappings = map[string]*Translation{
		"trafficManagement/proxyConfig/statusPort": {
			yAMLPath: "global/monitoringPort",
		},
	}
)

// defaultTranslationFunc is the default translation to values. It maps a Go data path into a YAML path.
func defaultTranslationFunc(t *Translation, root Tree, value string) error {
	setYAML(root, toPath(t.yAMLPath), value)
	return nil
}

// ProtoToValues traverses the supplied InstallerSpec and returns a values.yaml translation from it.
func ProtoToValues(ii *v1alpha1.InstallerSpec) (string, error) {
	root := make(Tree)

	protoToValues(*ii, root, nil)

	y, err := yaml.Marshal(root)
	if err != nil {
		return "", err
	}
	return string(y), err
}

// protoToValues takes an interface which must be a struct and recursively iterates through all its fields.
// For each leaf, if looks for a mapping from the struct data path to the corresponding YAML path and if one is
// found, it calls the associated mapping function if one is defined to populate the values YAML path.
// If no mapping function is defined, it uses the default mapping function.
func protoToValues(structVal interface{}, root Tree, path []string) {
	structElems := reflect.ValueOf(structVal).Elem()

	for i := 0; i < structElems.NumField(); i++ {
		fieldName := structElems.Type().Field(i).Name
		fieldValue := structElems.Field(i).Interface()

		switch structElems.Type().Field(i).Type.Kind() {
		case reflect.Ptr:
			protoToValues(structElems.Field(i).Elem().Interface(), root, append(path, fieldName))
		case reflect.Slice:
			// TODO
		default: // Must be a scalar leaf
			// See if we have a special handler
			m := mappings[fromPath(path)]
			v := fmt.Sprint(fieldValue)
			switch {
			case m == nil:
				// Default is to insert value at the same path as the source.
				setYAML(root, path, v)
			case m.translationFunc == nil:
				// Use default translation which just maps to a different part of the tree.
				defaultTranslationFunc(m, root, v)
			default:
				// Use a custom translation function.
				m.translationFunc(m, root, v)
			}
		}
	}
}

// setYAML sets the YAML path in the given Tree to the given value, creating any required intermediate nodes.
func setYAML(root Tree, path []string, value string) {
	if len(path) == 1 {
		root[path[0]] = value
		return
	}
	if root[path[0]] == nil {
		root[path[0]] = make(Tree)
	}
	setYAML(root[path[0]].(Tree), path[1:], value)
}

// toPath converts a string path of form a/b/c to a string slice representation.
func toPath(path string) []string {
	path = filepath.Clean(path)
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	return strings.Split(path, "/")
}

// fromPath converts a string slice path representation of form ["a", "b", "c"] to a string representation like "a/b/c".
func fromPath(path []string) string {
	return strings.Join(path, "/")
}
