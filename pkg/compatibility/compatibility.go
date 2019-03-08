// Package compatibility defines translations from installer proto to values.yaml.
package compatibility

import (
	"fmt"
	"reflect"
	"github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
	"github.com/ostromart/istio-installer/pkg/util"
	"gopkg.in/yaml.v2"
)

// TranslationFunc maps a proto API path into a YAML values tree.
type TranslationFunc func(t *Translation, root util.Tree, value string) error

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
func defaultTranslationFunc(t *Translation, root util.Tree, value string) error {
	setYAML(root, util.FromString(t.yAMLPath), value)
	return nil
}

// ProtoToValues traverses the supplied InstallerSpec and returns a values.yaml translation from it.
func ProtoToValues(ii *v1alpha1.InstallerSpec) (string, error) {
	root := make(util.Tree)

	errs := protoToValues(*ii, root, nil)

	y, err := yaml.Marshal(root)
	if err != nil {
		return "", util.AppendErr(errs, err).ToError()
	}
	return string(y), errs.ToError()
}

// protoToValues takes an interface which must be a struct and recursively iterates through all its fields.
// For each leaf, if looks for a mapping from the struct data path to the corresponding YAML path and if one is
// found, it calls the associated mapping function if one is defined to populate the values YAML path.
// If no mapping function is defined, it uses the default mapping function.
func protoToValues(structVal interface{}, root util.Tree, path util.Path) (errs util.Errors) {
	if reflect.TypeOf(structVal).Kind() != reflect.Struct {
		return util.NewErrs(fmt.Errorf("protoToValues path %s, value: %v, expected struct, got %T", path, structVal, structVal))
	}
	structElems := reflect.ValueOf(structVal).Elem()

	for i := 0; i < structElems.NumField(); i++ {
		fieldName := structElems.Type().Field(i).Name
		fieldValue := structElems.Field(i)
		fieldValueI := fieldValue.Interface()

		switch structElems.Type().Field(i).Type.Kind() {
		case reflect.Map:
			errs = util.AppendErr(errs, fmt.Errorf("protoToValues unexpected map type at path %s, fieldName", path, fieldName))
		case reflect.Ptr:
			errs = util.AppendErrs(errs, protoToValues(structElems.Field(i).Elem().Interface(), root, append(path, fieldName)))
		case reflect.Slice:
			for i := 0; i < fieldValue.Len(); i++ {
				errs = util.AppendErrs(errs, protoToValues(fieldValue.Index(i).Elem().Interface(), root, path))
			}
		default: // Must be a scalar leaf
			// See if we have a special handler
			m := mappings[path.String()]
			v := fmt.Sprint(fieldValueI)
			switch {
			case m == nil:
				// Default is to insert value at the same path as the source.
				setYAML(root, path, v)
			case m.translationFunc == nil:
				// Use default translation which just maps to a different part of the tree.
				errs = util.AppendErr(errs, defaultTranslationFunc(m, root, v))
			default:
				// Use a custom translation function.
				errs = util.AppendErr(errs, m.translationFunc(m, root, v))
			}
		}
	}
	return errs
}

// setYAML sets the YAML path in the given Tree to the given value, creating any required intermediate nodes.
func setYAML(root util.Tree, path util.Path, value string) {
	if len(path) == 1 {
		root[path[0]] = value
		return
	}
	if root[path[0]] == nil {
		root[path[0]] = make(util.Tree)
	}
	setYAML(root[path[0]].(util.Tree), path[1:], value)
}
