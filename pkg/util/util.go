package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

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