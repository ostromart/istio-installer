// Package compatibility defines translations from installer proto to values.yaml.
package compatibility

import (
	"fmt"
	"reflect"
	"github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
	"github.com/ostromart/istio-installer/pkg/util"
	"gopkg.in/yaml.v2"
	//	"github.com/godebug/pretty"
)

// TranslationFunc maps a proto API path into a YAML values tree.
type TranslationFunc func(t *Translation, root util.Tree, value interface{}) error

// Translation is a mapping between a proto data path and a YAML values path.
type Translation struct {
	yAMLPath        string
	translationFunc TranslationFunc
}

var (
	defaultMappings = map[string]*Translation{
		"TrafficManagement/ProxyConfig/StatusPort": {
			yAMLPath: "global/monitoringPort",
		},
	}
)

// defaultTranslationFunc is the default translation to values. It maps a Go data path into a YAML path.
func defaultTranslationFunc(t *Translation, root util.Tree, value interface{}) error {
	return setYAML(root, util.PathFromString(t.yAMLPath), value)
}

// ProtoToValues traverses the supplied InstallerSpec and returns a values.yaml translation from it. Mappings defines
// a mapping set
func ProtoToValues(mappings map[string]*Translation, ii *v1alpha1.InstallerSpec) (string, error) {
	root := make(util.Tree)

	errs := protoToValues(mappings, ii, root, nil)

	if len(root) == 0 {
		return "", nil
	}

	y, err := yaml.Marshal(root)
	if err != nil {
		return "", util.AppendErr(errs, err).ToError()
	}

	return string(y), errs.ToError()
}

// protoToValues takes an interface which must be a struct ptr and recursively iterates through all its fields.
// For each leaf, if looks for a mapping from the struct data path to the corresponding YAML path and if one is
// found, it calls the associated mapping function if one is defined to populate the values YAML path.
// If no mapping function is defined, it uses the default mapping function.
func protoToValues(mappings map[string]*Translation, structPtr interface{}, root util.Tree, path util.Path) (errs util.Errors) {
	fmt.Printf("protoToValues with path %s, %v (%T)\n", path, structPtr, structPtr)
	if structPtr == nil {
		return nil
	}
	if reflect.TypeOf(structPtr).Kind() != reflect.Ptr {
		return util.NewErrs(fmt.Errorf("protoToValues path %s, value: %v, expected ptr, got %T", path, structPtr, structPtr))
	}
	structElems := reflect.ValueOf(structPtr).Elem()
	if reflect.TypeOf(structElems).Kind() != reflect.Struct {
		return util.NewErrs(fmt.Errorf("protoToValues path %s, value: %v, expected struct, got %T", path, structElems, structElems))
	}

	if util.IsNilOrInvalidValue(structElems) {
		return
	}

	for i := 0; i < structElems.NumField(); i++ {
		fieldName := structElems.Type().Field(i).Name
		fieldValue := structElems.Field(i)
		kind := structElems.Type().Field(i).Type.Kind()

		//fmt.Printf("Checking field %s : %v\n", fieldName, fieldValue)
		switch kind {
		case reflect.Slice:
			for i := 0; i < fieldValue.Len(); i++ {
				errs = util.AppendErrs(errs, protoToValues(mappings, fieldValue.Index(i).Elem().Interface(), root, path))
			}
		case reflect.Ptr:
			if util.IsNilOrInvalidValue(fieldValue.Elem()) {
				//fmt.Println("value is nil, skip")
				continue
			}
			path = append(path, fieldName)
			if fieldValue.Elem().Kind() == reflect.Struct {
				errs = util.AppendErrs(errs, protoToValues(mappings, fieldValue.Interface(), root, path))
			} else {
				// Must be a scalar leaf. See if we have a mapping.
				m := mappings[path.String()]
				v := fieldValue.Elem().Interface()
				switch {
				case m == nil:
					// Default is to insert value at the same path as the source.
					errs = util.AppendErr(errs, setYAML(root, path, v))
				case m.translationFunc == nil:
					// Use default translation which just maps to a different part of the tree.
					fmt.Printf("Using default translation for %s.\n", path)
					errs = util.AppendErr(errs, defaultTranslationFunc(m, root, v))
				default:
					// Use a custom translation function.
					errs = util.AppendErr(errs, m.translationFunc(m, root, v))
				}
			}
		}
	}
	return errs
}

// setYAML sets the YAML path in the given Tree to the given value, creating any required intermediate nodes.
func setYAML(root util.Tree, path util.Path, value interface{}) error {
	//fmt.Printf("setYAML %s:%s\n", path, value)
	if len(path) == 0 {
		return fmt.Errorf("path cannot be empty")
	}
	if len(path) == 1 {
		root[path[0]] = value
		return nil
	}
	if root[path[0]] == nil {
		root[path[0]] = make(util.Tree)
	}
	setYAML(root[path[0]].(util.Tree), path[1:], value)
	return nil
}
