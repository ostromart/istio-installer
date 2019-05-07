package patch

import (
	"context"
	"fmt"
	"github.com/kr/pretty"
	"github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
	"github.com/ostromart/istio-installer/pkg/manifest"
	"github.com/ostromart/istio-installer/pkg/util"
	"gopkg.in/yaml.v2"
	"istio.io/istio/pkg/log"
	"strings"
)

func PatchYAMLManifest(baseYAML string, namespace string, resourceOverride []*v1alpha1.K8SObjectOverlay) (string, error) {
	baseObjs, err := manifest.ParseObjectsFromYAMLManifest(context.TODO(), baseYAML)
	if err != nil {
		return "", err
	}

	bom := baseObjs.ToMap()
	oom, err := objectOverrideMap(resourceOverride, namespace)
	if err != nil {
		return "", err
	}
	fmt.Println(bom)
	fmt.Println(oom)
	var ret strings.Builder

	// Try to apply the defined overlays.
	for k, oo := range oom {
		bo := bom[k]
		if bo == nil {
			// TODO: error log overlays with no matches in any component.
			continue
		}
		patched, err := applyPatches(bo, oo)
		if err != nil {
			log.Errorf("patch error: %s", err)
			continue
		}
		//log.Infof("Base object: \n%s\nAfter overlay:\n%s", bo.YAMLDebugString(), patched)
		ret.Write(patched)
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

func applyPatches(base *manifest.Object, patches []*v1alpha1.K8SObjectOverlay_PathValue) (outYAML []byte, errs util.Errors) {
	bo := make(map[interface{}]interface{})
	by, err := base.YAML()
	if err != nil {
		return nil, util.NewErrs(err)
	}
	err = yaml.Unmarshal(by, bo)
	if err != nil {
		return nil, util.NewErrs(err)
	}
	for _, p := range patches {
		fmt.Printf("applying path=%s, value=%s\n", p.Path, p.Value)
		inc, err := getNode(makeNodeContext(bo), util.PathFromString(p.Path))
		if err != nil {
			fmt.Println(err)
			errs = util.AppendErr(errs, err)
			continue
		}
		fmt.Printf("before delete(%p):\n%s\n", inc.parent, pretty.Sprint(inc.parent))
		errs = util.AppendErr(errs, writeNode(inc, p.Value))
		fmt.Printf("%p\n", bo["spec"].(map[interface{}]interface{})["template"].(map[interface{}]interface{})["spec"].(map[interface{}]interface{})["containers"].([]interface{})[1].(map[interface{}]interface{})["command"])
		fmt.Printf("after delete(%p)...\n%s\n%s\n", inc.parent, pretty.Sprint(bo), pretty.Sprint(inc.parent))
	}
	oy, err := yaml.Marshal(bo)
	if err != nil {
		return nil, util.AppendErr(errs, err)
	}
	return oy, errs
}

type nodeContext struct {
	grandparent interface{}
	parent      interface{}
	node        interface{}
	parentKey   interface{}
	key         interface{}
	index       int
}

func (nc *nodeContext) String() string {
	return pretty.Sprint(*nc)
}

func makeNodeContext(obj interface{}) *nodeContext {
	return &nodeContext{
		node:  obj,
		index: -1,
	}
}

func getNode(nc *nodeContext, path util.Path) (*nodeContext, error) {
	if len(path) == 0 {
		fmt.Printf("path end nodeContext=\n%s\n", nc)
		if util.IsMap(nc.node) {
			ret := *nc
			ret.parent = nc.node
			return &ret, nil
		}
		return nc, nil
	}

	pe := path[0]
	fmt.Printf("getNode path=%s, nodeContext=%s\n", path, nc)

	// list or leaf list
	if lst, ok := nc.node.([]interface{}); ok {
		fmt.Println("list type")
		for idx, le := range lst {
			fmt.Printf("idx=%d\n", idx)
			if lm, ok := le.(map[interface{}]interface{}); ok {
				fmt.Println("node list")
				k, v, err := pathKV(pe)
				if err != nil {
					return nil, err
				}
				if stringsEqual(lm[k], v) {
					fmt.Printf("found matching kv %v:%v\n", k, v)
					nnc := &nodeContext{
						grandparent: nc.parent,
						parent:      &nc.node,
						node:        lm,
						parentKey:   k,
						key:         k,
						index:       idx,
					}
					return getNode(nnc, path[1:])
				}
				continue
			}
			// Must be a leaf list
			fmt.Println("leaf list")
			v, err := pathV(pe)
			if err != nil {
				return nil, err
			}
			if stringsEqual(v, le) {
				fmt.Printf("found matching key %v\n", le)
				nnc := &nodeContext{
					grandparent: nc.parent,
					parent:      &nc.node,
					node:        le,
					parentKey:   nc.key,
					key:         v,
					index:       idx,
				}
				return getNode(nnc, path[1:])
			}
		}
		return nil, fmt.Errorf("path element %s not found", pe)
	}

	// interior or non-leaf node
	if nn, ok := nc.node.(map[interface{}]interface{}); ok {
		nnc := &nodeContext{
			grandparent: nc.parent,
			parent:      &nc.node,
			node:        nn[pe],
			parentKey:   pe,
			key:         pe,
			index:       -1,
		}
		return getNode(nnc, path[1:])
	}

	return nil, fmt.Errorf("leaf type %T in non-leaf node %s", nc.node, path)
}

func writeNode(nc *nodeContext, value interface{}) error {
	fmt.Printf("writeNode nodeContext=\n%s\n, value=%v\n", nc, value)
	if nc.parent == nil {
		nc.key = value
		return nil
	}

	if util.IsPtr(nc.parent) {
		// must be a slice with map parent
		//l := reflect.ValueOf(parent).Elem().Interface()
		fmt.Printf("list index %d\n", nc.index)
		if nc.index == -1 {
			fmt.Println("append")
			if err := util.AppendToSlicePtr(nc.parent, value); err != nil {
				return err
			}
		}
		if value == nil {
			fmt.Println("delete")
			if err := util.DeleteFromSlicePtr(nc.parent, nc.index); err != nil {
				return err
			}
		} else {
			fmt.Println("update")
			if err := util.UpdateSlicePtr(nc.parent, nc.index, value); err != nil {
				return err
			}
		}

		if err := util.InsertIntoMap(nc.grandparent, nc.parentKey, nc.parent); err != nil {
			return err
		}

		return nil
	}
	if pmap, ok := nc.parent.(map[interface{}]interface{}); ok {
		fmt.Println("map")
		pmap[nc.key] = value
		return nil
	}
	fmt.Println("leaf")
	nc.node = value
	return nil
}

func isValidPathElement(pe string) bool {
	return util.ValidKeyRegex.MatchString(pe)
}

func isKVPathElement(pe string) bool {
	if len(pe) < 3 {
		return false
	}
	kv := strings.Split(pe, ":")
	if len(kv) != 2 {
		return false
	}
	return isValidPathElement(kv[0])
}

func isVPathElement(pe string) bool {
	return len(pe) > 1 && pe[0] == ':'
}

func pathV(pe string) (string, error) {
	if !isVPathElement(pe) {
		return "", fmt.Errorf("%s is not a valid value path element", pe)
	}
	return pe[1:], nil
}

func pathKV(pe string) (k, v string, err error) {
	if !isKVPathElement(pe) {
		return "", "", fmt.Errorf("%s is not a valid key:value path element", pe)
	}
	kv := strings.Split(pe, ":")
	return kv[0], kv[1], nil
}

func objectOverrideMap(oos []*v1alpha1.K8SObjectOverlay, namespace string) (map[string][]*v1alpha1.K8SObjectOverlay_PathValue, error) {
	ret := make(map[string][]*v1alpha1.K8SObjectOverlay_PathValue)
	for _, o := range oos {
		ret[manifest.Hash(o.Kind, namespace, o.Name)] = o.Patches
	}
	return ret, nil
}

func stringsEqual(a, b interface{}) bool {
	return fmt.Sprint(a) == fmt.Sprint(b)
}
