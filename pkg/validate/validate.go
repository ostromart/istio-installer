package validate

import (
	"fmt"
	"reflect"

	"github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
	"github.com/ostromart/istio-installer/pkg/util"
)

// ValidateFunc validates a value.
type ValidateFunc func(i interface{}) error

var (
	// validateFuncs maps a data path to a validation function.
	validateFuncs = map[string]ValidateFunc{
		"trafficManagement/proxyConfig/statusPort": trafficManagementProxyConfigStatusPortValidateFunc,
	}
)

func trafficManagementProxyConfigStatusPortValidateFunc(i interface{}) error {
	return validatePortNumber(i)
}

// Validate validates the values in the given Installer spec, using the field map validateFuncs to call the appropriate
// validation function.
func Validate(is *v1alpha1.InstallerSpec) (errs []error) {
	return validate(is, nil)
}

// validate takes a struct type and recurses through its fields until it reaches a leaf. It validates each leaf
// according to the validation function for its path if one is defined in validateFuncs if one is found, or skips the
// leaf otherwise. All errors encountered during validation are returned as a slice.
func validate(structVal interface{}, path util.Path) (errs []error) {
	if reflect.TypeOf(structVal).Kind() != reflect.Struct {
		return util.NewErrs(fmt.Errorf("protoToValues value: %v, expected struct, got %T", path, structVal, structVal))
	}
	structElems := reflect.ValueOf(structVal).Elem()

	for i := 0; i < structElems.NumField(); i++ {
		fieldName := structElems.Type().Field(i).Name
		fieldValue := structElems.Field(i)

		switch structElems.Type().Field(i).Type.Kind() {
		case reflect.Map:
			errs = util.AppendErr(errs, fmt.Errorf("protoToValues unexpected map type at path %s, fieldName", path, fieldName))
		case reflect.Ptr:
			errs = util.AppendErrs(errs, validate(structElems.Field(i).Elem().Interface(), append(path, fieldName)))
		case reflect.Slice:
			for i := 0; i < fieldValue.Len(); i++ {
				errs = util.AppendErrs(errs, validate(structElems.Field(i).Elem().Interface(), path))
			}
		default: // Must be a scalar leaf
			if m := validateFuncs[path.String()]; m != nil {
				errs = util.AppendErr(errs, m(structElems.Field(i).Elem().Interface()))
			}
		}
	}
	return errs
}

func validatePortNumber(i interface{}) error {
	return validateIntRange(i, 0, 65535)
}

func validateIntRange(i interface{}, min, max int64) error {
	k := reflect.TypeOf(i).Kind()
	switch {
	case isIntKind(k):
		v := reflect.ValueOf(i).Int()
		if v < min || v > max {
			return fmt.Errorf("value %v falls out side range [%v, %v]", v, min, max)
		}
	case isUintKind(k):
		v := reflect.ValueOf(i).Uint()
		if int64(v) < min || int64(v) > max {
			return fmt.Errorf("value %v falls out side range [%v, %v]", v, min, max)
		}
	}

	return nil
}

func isIntKind(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	}
	return false
}

func isUintKind(k reflect.Kind) bool {
	switch k {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	}
	return false
}
