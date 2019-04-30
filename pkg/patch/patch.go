package patch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/evanphx/json-patch"
	"github.com/ghodss/yaml"
	"github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
	"github.com/ostromart/istio-installer/pkg/manifest"
	"github.com/ostromart/istio-installer/pkg/util"
	"istio.io/istio/pkg/log"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"strings"
)

func PatchYAMLManifest(baseYAML string, resourceOverride []*v1alpha1.K8SObjectOverlay) (string, error) {
	baseObjs, err := manifest.ParseObjectsFromYAMLManifest(context.TODO(), baseYAML)
	if err != nil {
		return "", err
	}

	bom := baseObjs.ToMap()
	oom, err := objectOverrideMap(resourceOverride)
	if err != nil {
		return "", err
	}
	var ret strings.Builder

	// Try to apply the defined overlays.
	for k, oo := range oom {
		bo := bom[k]
		if bo == nil {
			// TODO: error log overlays with no matches in any component.
			continue
		}
		boj, err := bo.JSON()
		if err != nil {
			log.Errorf("JSON error (%s) for base manifest object: \n%v", err, bo)
			continue
		}
		merged, err := applyPatch(boj, oo)
		if err != nil {
			log.Errorf("JSON merge error: %s", err)
			continue
		}
		//log.Infof("Base object: \n%s\nAfter overlay:\n%s", bo.YAMLDebugString(), merged)
		ret.Write(merged)
		ret.WriteString("\n---\n")
	}
	// Render the remaining objects with no overlays.
	for k, oo := range bom {
		if oom[k] != nil {
			// Skip objects that have overlays.
			continue
		}
		oy, err := oo.YAML()
		if err != nil {
			log.Errorf("Object to YAML error (%s) for base object: \n%v", err, oo)
			continue
		}
		ret.Write(oy)
		ret.WriteString("\n---\n")
	}
	return ret.String(), nil
}

func getNode(parent, node interface{}, path util.Path, createNodes bool) (key, keyParent interface{}, err error) {
	if len(path) == 0 {
		return parent, node, nil
	}

	pe := path[0]
	fmt.Printf("%s: ", pe)
	switch {
	case isSimplePathElement(pe):
		fmt.Println("simple")
		nn, ok := node.(map[string]interface{})
		if !ok {
			return nil, nil, fmt.Errorf("path element %s node has type %T, expecting map", pe, node)
		}
		if nn[pe] == nil {
			if !createNodes {
				return nil, nil, fmt.Errorf("path element %s is missing and not creating missing nodes", pe)
			}
			nn[pe] = make(map[string]interface{})
		}
		return getNode(nn, nn[pe], path[1:], createNodes)

	case isKVPathElement(pe):
		fmt.Println("kv")
		k, v, err := pathKV(pe)
		if err != nil {
			return nil, nil, err
		}
		nn, ok := node.([]interface{})
		if !ok {
			return nil, nil, fmt.Errorf("path element %s node has type %T, expecting slice", pe, node)
		}
		for i, ni := range nn {
			n, ok := ni.(map[string]interface{})
			if !ok {
				return nil, nil, fmt.Errorf("list element path %s has type %T, expecting map", pe, ni)
			}
			fmt.Printf("check %v == %v? %v\n", n[k], v)
			if fmt.Sprint(n[k]) == fmt.Sprint(v) {
				return n, i, nil
			}
		}
		// No list elements match.
		if !createNodes {
			return nil, nil, fmt.Errorf("path element %s is missing and not creating missing nodes", pe)
		}
		ne := map[string]interface{}{
			k: v,
		}
		nn = append(nn, ne)
		return getNode(nn, ne, path[1:], createNodes)

	case isVPathElement(pe):
		fmt.Println("v")
		// Leaf-list
		v, err := pathV(pe)
		if err != nil {
			return nil, nil, err
		}
		nl, ok := node.([]interface{})
		if !ok {
			return nil, nil, fmt.Errorf("path element %s node has type %T, expecting slice", pe, node)
		}
		if len(path) != 1 {
			return nil, nil, fmt.Errorf("path element %s is a list leaf, but path %s len > 1", pe, path)
		}
		for i, n := range nl {
			if n == v {
				return nl, i, nil
			}
		}

	case isKKVPathElement(pe):
		fmt.Println("kkv")
		k, kv2, err := pathKKV(pe)
		if err != nil {
			return nil, nil, err
		}
		nn, ok := node.(map[string]interface{})
		if !ok {
			return nil, nil, fmt.Errorf("path element %s node has type %T, expecting map", pe, node)
		}
		path[0] = kv2
		return getNode(nn, nn[k], path, createNodes)

	default:
		return nil, nil, fmt.Errorf("bad path element %s", pe)
	}

	return
}

func isValidPathElement(pe string) bool {
	return util.ValidKeyRegex.MatchString(pe)
}

func isSimplePathElement(pe string) bool {
	return isValidPathElement(pe)
}

func isKVPathElement(pe string) bool {
	return len(pe) > 0 && pe[0] == '[' && pe[len(pe)-1] == ']' && strings.Contains(pe, ":")
}

func isVPathElement(pe string) bool {
	return len(pe) > 0 && pe[0] == '[' && pe[len(pe)-1] == ']' && !strings.Contains(pe, ":")
}

func isKKVPathElement(pe string) bool {
	return len(pe) > 0 && strings.Contains(pe, "[") && strings.Contains(pe, "]") && strings.Contains(pe, ":")
}

func pathKKV(pe string) (k1, kv string, err error) {
	i := strings.Index(pe, "[")
	if i <= 0 {
		return "", "", fmt.Errorf("path element %s does not have expected form [key:value]", pe)
	}
	return pe[0:i], pe[i:], nil
}

func pathV(pe string) (string, error) {
	if !isVPathElement(pe) {
		return "", fmt.Errorf("%s is not a valid value path element", pe)
	}
	pn := pe[1 : len(pe)-1]
	if !isValidPathElement(pn) {
		return "", fmt.Errorf("illegal value path element %s", pe)
	}
	return pn, nil
}

func pathKV(pe string) (k, v string, err error) {
	if pe == "" {
		return "", "", errors.New("KV path element must not be empty")
	}
	if pe[0] != '[' {
		return "", "", fmt.Errorf("KV path element %s must begin with [", pe)
	}
	if pe[len(pe)-1] != ']' {
		return "", "", fmt.Errorf("KV path element %s must end with ]", pe)
	}
	kv := strings.Split(pe[1:len(pe)-1], ":")
	if len(kv) != 2 {
		return "", "", fmt.Errorf("KV path element %s must have the form [key:value]", pe)
	}
	k, v = kv[0], kv[1]
	if !isValidPathElement(k) {
		return "", "", fmt.Errorf("KV path element %s has bad key %s", pe, k)
	}
	return k, v, nil
}

// setYAML sets the YAML path in the given Tree to the given value, creating any required intermediate nodes.
func setYAML(root util.Tree, path util.Path, value interface{}) error {
	fmt.Printf("setYAML %s:%v\n", path, value)
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

func applyPatch(boj []byte, oo *v1alpha1.K8SObjectOverlay) ([]byte, error) {
	switch oo.PatchType {
	case v1alpha1.K8SObjectOverlay_JSON:
	default:
		return nil, fmt.Errorf("Unsupported patch type %v", oo.PatchType)
	}
	switch oo.Op {
	case v1alpha1.K8SObjectOverlay_MERGE:
	default:
		return nil, fmt.Errorf("Unsupported patch op %v", oo.Op)
	}

	ooj, err := json.Marshal(oo.Data)
	merged, err := jsonpatch.MergePatch(boj, ooj)
	if err != nil {
		return nil, fmt.Errorf("JSON merge error (%s) for base object: \n%s\n override object: \n%s", err, boj, ooj)
	}
	my, err := yaml.JSONToYAML(merged)
	if err != nil {
		return nil, fmt.Errorf("JSONToYAML error (%s) for merged object: \n%s", err, merged)
	}
	return my, nil
}

func interfaceMapToUnstructured(m map[string]interface{}) (*unstructured.Unstructured, error) {
	jb, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	out := &unstructured.Unstructured{}
	err = json.Unmarshal(jb, out)
	return out, err
}

func objectOverrideMap(oos []*v1alpha1.K8SObjectOverlay) (map[string]*v1alpha1.K8SObjectOverlay, error) {
	ret := make(map[string]*v1alpha1.K8SObjectOverlay)
	for _, o := range oos {
		u, err := interfaceMapToUnstructured(o.Data)
		if err != nil {
			return nil, fmt.Errorf("interfaceMapToUnstructured with %v: %v", o.Data, err)
		}
		ret[manifest.Hash(u)] = o
	}
	return ret, nil
}
