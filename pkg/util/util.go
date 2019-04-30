package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"regexp"
	"strings"
	"math/rand"
	"reflect"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	ValidKeyRegex = regexp.MustCompile("^[a-zA-Z0-9_-]*$")
)

func GetPathVal(tree map[string]interface{}, path string) (string, bool) {
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	pv := strings.Split(path, "/")

	for ; len(pv) > 0; pv = pv[1:] {
		p := pv[0]
		v, ok := tree[p]
		if !ok {
			return "", false
		}
		if len(pv) == 1 {
			return fmt.Sprint(v), true
		}
		tree, ok = v.(map[string]interface{})
		if !ok {
			return "", false
		}
	}

	return "", false
}

func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func PrettyJSON(b []byte) []byte {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	if err != nil {
		return []byte(fmt.Sprint(err))
	}
	return out.Bytes()
}

// IsValueNil returns true if either value is nil, or has dynamic type {ptr,
// map, slice} with value nil.
func IsValueNil(value interface{}) bool {
	if value == nil {
		return true
	}
	switch reflect.TypeOf(value).Kind() {
	case reflect.Slice, reflect.Ptr, reflect.Map:
		return reflect.ValueOf(value).IsNil()
	}
	return false
}

// IsEmptyString returns true if value is an empty string.
func IsEmptyString(value interface{}) bool {
	if value == nil {
		return true
	}
	switch reflect.TypeOf(value).Kind() {
	case reflect.String:
		return value.(string) == ""
	}
	return false
}

// IsEmptyString returns true if value is an empty string.
func IsSlice(value interface{}) bool {
	return reflect.TypeOf(value).Kind() == reflect.Slice
}

// IsNilOrInvalidValue reports whether v is nil or reflect.Zero.
func IsNilOrInvalidValue(v reflect.Value) bool {
	return !v.IsValid() || (v.Kind() == reflect.Ptr && v.IsNil()) || IsValueNil(v.Interface())
}


func IsYAMLEqual(a, b string) bool {
	if strings.TrimSpace(a) == "" && strings.TrimSpace(b) == "" {
		return true
	}
	ajb, err := yaml.YAMLToJSON([]byte(a))
	if err != nil {
		fmt.Printf("bad YAML in isYAMLEqual:\n%s\n", a)
		return false
	}
	bjb, err := yaml.YAMLToJSON([]byte(b))
	if err != nil {
		fmt.Printf("bad YAML in isYAMLEqual:\n%s\n", b)
		return false
	}

	//fmt.Printf("a:\n%s\nb:\n%s\n", string(ajb), string(bjb))
	return string(ajb) == string(bjb)
}