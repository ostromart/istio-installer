package validate

import (
	"fmt"
	"net"
	"reflect"
	"strings"

	"github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
	"github.com/ostromart/istio-installer/pkg/util"
)

// ValidateFunc validates a value.
type ValidateFunc func(i interface{}) error

var (
	// defaultValidations maps a data path to a validation function.
	defaultValidations = map[string]ValidateFunc{
		"TrafficManagement/IncludeIpRanges":  validateStringList(validateCIDR),
		"TrafficManagement/ExcludeIpRanges":  validateStringList(validateCIDR),
		"TrafficManagement/IncludeInboundPorts":  validateStringList(validatePortNumber),
		"TrafficManagement/IncludeOutboundPorts":  validateStringList(validatePortNumber),
	}
)

// Validate validates the values in the given Installer spec, using the field map defaultValidations to call the appropriate
// validation function.
func Validate(validations map[string]ValidateFunc, is *v1alpha1.InstallerSpec) error {
	return validate(defaultValidations, is, nil).ToError()
}

func validate(validations map[string]ValidateFunc, structPtr interface{}, path util.Path) (errs util.Errors) {
	//fmt.Printf("validate with path %s, %v (%T)\n", path, structPtr, structPtr)
	if structPtr == nil {
		return nil
	}
	if reflect.TypeOf(structPtr).Kind() != reflect.Ptr {
		return util.NewErrs(fmt.Errorf("validate path %s, value: %v, expected ptr, got %T", path, structPtr, structPtr))
	}
	structElems := reflect.ValueOf(structPtr).Elem()
	if reflect.TypeOf(structElems).Kind() != reflect.Struct {
		return util.NewErrs(fmt.Errorf("validate path %s, value: %v, expected struct, got %T", path, structElems, structElems))
	}

	if util.IsNilOrInvalidValue(structElems) {
		return
	}

	for i := 0; i < structElems.NumField(); i++ {
		fieldName := structElems.Type().Field(i).Name
		fieldValue := structElems.Field(i)
		kind := structElems.Type().Field(i).Type.Kind()
		if a, ok := structElems.Type().Field(i).Tag.Lookup("json"); ok && a == "-" {
			continue
		}

		//fmt.Printf("Checking field %s\n", fieldName)
		switch kind {
		case reflect.Struct:
			//fmt.Println("Struct")
			errs = util.AppendErrs(errs, validate(validations, fieldValue.Addr().Interface(), append(path, fieldName)))
		case reflect.Map:
			//fmt.Println("Map")
			newPath := append(path, fieldName)
			for _, key := range fieldValue.MapKeys() {
				nnp := append(newPath, key.String())
				errs = util.AppendErr(errs, validateLeaf(validations, nnp, fieldValue.MapIndex(key)))
			}
		case reflect.Slice:
			//fmt.Println("Slice")
			for i := 0; i < fieldValue.Len(); i++ {
				errs = util.AppendErrs(errs, validate(validations, fieldValue.Index(i).Elem().Interface(), path))
			}
		case reflect.Ptr:
			if util.IsNilOrInvalidValue(fieldValue.Elem()) {
				//fmt.Println("value is nil, skip")
				continue
			}
			newPath := append(path, fieldName)
			if fieldValue.Elem().Kind() == reflect.Struct {
				//fmt.Println("Struct Ptr")
				errs = util.AppendErrs(errs, validate(validations, fieldValue.Interface(), newPath))
			} else {
				//fmt.Println("Leaf Ptr")
				errs = util.AppendErr(errs, validateLeaf(validations, newPath, fieldValue))
			}
		default:
			//fmt.Printf("field has kind %s\n", kind)
			if structElems.Field(i).CanInterface() {
				errs = util.AppendErr(errs, validateLeaf(validations, append(path, fieldName), fieldValue.Interface()))
			}
		}
	}
	return errs
}

func validateLeaf(validations map[string]ValidateFunc, path util.Path, val interface{}) error {
	fmt.Printf("validate %s:%v(%T) ", path.String(), val, val)
	if util.IsValueNil(val) {
		// TODO(mostrowski): handle required fields.
		fmt.Printf("validate %s: OK (nil value)\n", path.String())
		return nil
	}
	vf, ok := validations[path.String()]
	if !ok {
		fmt.Printf("validate %s: OK (no validation)\n", path.String())
		// No validation defined.
		return nil
	}
	return vf(val)
}

func validatePortNumber(val interface{}) error {
	return validateIntRange(val, 0, 65535)
}

func validateIntRange(val interface{}, min, max int64) error {
	fmt.Printf("validateIntRange %v in [%d, %d]?: ", val, min, max)
	k := reflect.TypeOf(val).Kind()
	var err error
	switch {
	case isIntKind(k):
		v := reflect.ValueOf(val).Int()
		if v < min || v > max {
			err = fmt.Errorf("value %v falls out side range [%v, %v]", v, min, max)
		}
	case isUintKind(k):
		v := reflect.ValueOf(val).Uint()
		if int64(v) < min || int64(v) > max {
			err = fmt.Errorf("value %v falls out side range [%v, %v]", v, min, max)
		}
	}
	printError(err)
	return err
}

func validateCIDR(val interface{}) error {
	fmt.Printf("validateCIDR (%s): ", val)
	var err error
	if reflect.TypeOf(val).Kind() != reflect.String {
		err = fmt.Errorf("validateCIDR got %T, want string", val)
	} else {
		_, _, err = net.ParseCIDR(val.(string))
	}
	printError(err)
	return err
}

func validateStringList(vf ValidateFunc) ValidateFunc {
	return func(val interface{}) error {
		fmt.Printf("validateStringList(\n")
		if reflect.TypeOf(val).Kind() != reflect.String {
			err := fmt.Errorf("validateStringList got %T, want string", val)
			printError(err)
			return err
		}
		var errs util.Errors
		for _, s := range strings.Split(val.(string), ",") {
			errs = util.AppendErr(errs, vf(strings.TrimSpace(s)))
		}
		err := errs.ToError()
		fmt.Print("):")
		printError(err)
		return err
	}
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

func printError(err error) {
	if err == nil {
		fmt.Println("OK")
		return
	}
	fmt.Println(err)
}

