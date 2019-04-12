package validate

import (
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"

	"github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
	"github.com/ostromart/istio-installer/pkg/util"
)

// ValidateFunc validates a value.
type ValidateFunc func(path util.Path, i interface{}) error

var (
	// defaultValidations maps a data path to a validation function.
	defaultValidations = map[string]ValidateFunc{
		"TrafficManagement/IncludeIpRanges":  validateStringList(validateCIDR),
		"TrafficManagement/ExcludeIpRanges":  validateStringList(validateCIDR),
		"TrafficManagement/IncludeInboundPorts":  validateStringList(validatePortNumberString),
		"TrafficManagement/ExcludeInboundPorts":  validateStringList(validatePortNumberString),
	}

	// requiredValues lists all the values that must be non-empty.
	requiredValues = map[string]bool {
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
	pstr := path.String()
	fmt.Printf("validate %s:%v(%T) ", pstr, val, val)
	if !requiredValues[pstr] && (util.IsValueNil(val) || util.IsEmptyString(val)) {
		// TODO(mostrowski): handle required fields.
		fmt.Printf("validate %s: OK (empty value)\n", pstr)
		return nil
	}

	vf, ok := validations[pstr]
	if !ok {
		fmt.Printf("validate %s: OK (no validation)\n", pstr)
		// No validation defined.
		return nil
	}
	return vf(path, val)
}

func validatePortNumberString(path util.Path, val interface{}) error {
	fmt.Printf("validatePortNumberString %v: ", val)
	if !isString(val) {
		return fmt.Errorf("validatePortNumberString(%s) bad type %T, want string", path, val)
	}
	intV, err := strconv.ParseInt(val.(string), 10, 32);
	if err != nil {
		return fmt.Errorf("%s : %s", path, err)
	}
	return validatePortNumber(path, intV)
}

func validatePortNumber(path util.Path, val interface{}) error {
	return validateIntRange(path, val, 0, 65535)
}

func validateIntRange(path util.Path, val interface{}, min, max int64) error {
	fmt.Printf("validateIntRange %s:%v in [%d, %d]?: ", path, val, min, max)
	k := reflect.TypeOf(val).Kind()
	var err error
	switch {
	case isIntKind(k):
		v := reflect.ValueOf(val).Int()
		if v < min || v > max {
			err = fmt.Errorf("value %s:%v falls outside range [%v, %v]", path, v, min, max)
		}
	case isUintKind(k):
		v := reflect.ValueOf(val).Uint()
		if int64(v) < min || int64(v) > max {
			err = fmt.Errorf("value %s:%v falls out side range [%v, %v]", path, v, min, max)
		}
	default:
		err = fmt.Errorf("validateIntRange %s unexpected type %T, want int type", path, val)
	}
	printError(err)
	return err
}

func validateCIDR(path util.Path, val interface{}) error {
	fmt.Printf("validateCIDR (%s): ", val)
	var err error
	if reflect.TypeOf(val).Kind() != reflect.String {
		err = fmt.Errorf("validateCIDR %s got %T, want string", path, val)
	} else {
		_, _, err = net.ParseCIDR(val.(string))
		if err != nil {
			err = fmt.Errorf("%s %s", path, err)
		}
	}
	printError(err)
	return err
}

func validateStringList(vf ValidateFunc) ValidateFunc {
	return func(path util.Path, val interface{}) error {
		fmt.Printf("validateStringList(\n")
		if reflect.TypeOf(val).Kind() != reflect.String {
			err := fmt.Errorf("validateStringList %s got %T, want string", path, val)
			printError(err)
			return err
		}
		var errs util.Errors
		for _, s := range strings.Split(val.(string), ",") {
			errs = util.AppendErr(errs, vf(path, strings.TrimSpace(s)))
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

func isString(val interface{}) bool {
	return reflect.TypeOf(val).Kind() == reflect.String
}

func printError(err error) {
	if err == nil {
		fmt.Println("OK")
		return
	}
	fmt.Println(err)
}

