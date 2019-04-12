package validate

import (
	"fmt"
	"github.com/ostromart/istio-installer/pkg/util"
	"net"
	"reflect"
	"strconv"
	"strings"
)

// ValidateFunc validates a value.
type ValidateFunc func(path util.Path, i interface{}) util.Errors

func validateStringList(vf ValidateFunc) ValidateFunc {
	return func(path util.Path, val interface{}) util.Errors {
		fmt.Printf("validateStringList(\n")
		if reflect.TypeOf(val).Kind() != reflect.String {
			err := fmt.Errorf("validateStringList %s got %T, want string", path, val)
			printError(err)
			return util.NewErrs(err)
		}
		var errs util.Errors
		for _, s := range strings.Split(val.(string), ",") {
			errs = util.AppendErrs(errs, vf(path, strings.TrimSpace(s)))
			fmt.Printf("\nerrors(%d): %v\n", len(errs), errs)
		}
		printError(errs.ToError())
		return errs
	}
}

func validatePortNumberString(path util.Path, val interface{}) util.Errors {
	fmt.Printf("validatePortNumberString %v: ", val)
	if !isString(val) {
		return util.NewErrs(fmt.Errorf("validatePortNumberString(%s) bad type %T, want string", path, val))
	}
	intV, err := strconv.ParseInt(val.(string), 10, 32)
	if err != nil {
		return util.NewErrs(fmt.Errorf("%s : %s", path, err))
	}
	return validatePortNumber(path, intV)
}

func validatePortNumber(path util.Path, val interface{}) util.Errors {
	return validateIntRange(path, val, 0, 65535)
}

func validateIntRange(path util.Path, val interface{}, min, max int64) util.Errors {
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
	return util.NewErrs(err)
}

func validateCIDR(path util.Path, val interface{}) util.Errors {
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
	return util.NewErrs(err)
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
