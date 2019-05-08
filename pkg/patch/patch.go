package patch

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/kr/pretty"
	"github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
	"github.com/ostromart/istio-installer/pkg/manifest"
	"github.com/ostromart/istio-installer/pkg/util"
	"gopkg.in/yaml.v2"
	"istio.io/istio/pkg/log"
)

var (
	// debugPackage controls verbose debugging in this package. Used for offline debugging.
	debugPackage = false
)

// nodeContext provides a means for traversing a tree towards the root.
type nodeContext struct {
	// parent in the parent of this nodeContext
	parent *nodeContext
	// keyToChild is the key required to reach the child.
	keyToChild interface{}
	// node is the actual node in the data tree
	node interface{}
}

// String implements the Stringer interface.
func (nc *nodeContext) String() string {
	ret := "\n--------------- NodeContext ------------------\n"
	ret += fmt.Sprintf("parent.node=\n%s\n", pretty.Sprint(nc.parent.node))
	ret += fmt.Sprintf("keyToChild=%v\n", nc.parent.keyToChild)
	ret += fmt.Sprintf("node=\n%s\n", pretty.Sprint(nc.node))
	ret += "----------------------------------------------\n"
	return ret
}

// makeNodeContext returns a nodeContext created from the given object.
func makeNodeContext(obj interface{}) *nodeContext {
	return &nodeContext{
		node: obj,
	}
}

// PatchYAMLManifest patches a base YAML in the given namespace with a list of overlays.
// Each overlay has the format described in the K8SObjectOverlay definition.
// It returns the patched manifest YAML.
func PatchYAMLManifest(baseYAML string, namespace string, overlays []*v1alpha1.K8SObjectOverlay) (string, error) {
	baseObjs, err := manifest.ParseObjectsFromYAMLManifest(context.TODO(), baseYAML)
	if err != nil {
		return "", err
	}

	bom := baseObjs.ToMap()
	oom, err := objectOverrideMap(overlays, namespace)
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
		patched, err := applyPatches(bo, oo)
		if err != nil {
			log.Errorf("patch error: %s", err)
			continue
		}
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

// applyPatches applies the given patches against the given object. It returns the resulting patched YAML if successful,
// or a list of errors otherwise.
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
		dbgPrint("applying path=%s, value=%s\n", p.Path, p.Value)
		inc, err := getNode(makeNodeContext(bo), util.PathFromString(p.Path))
		if err != nil {
			fmt.Println(err)
			errs = util.AppendErr(errs, err)
			continue
		}
		errs = util.AppendErr(errs, writeNode(inc, p.Value))
	}
	oy, err := yaml.Marshal(bo)
	if err != nil {
		return nil, util.AppendErr(errs, err)
	}
	return oy, errs
}

// getNode returns the node which has the given patch from the source node given by nc.
// It creates a tree of nodeContexts during the traversal so that parent structures can be updated if required.
func getNode(nc *nodeContext, path util.Path) (*nodeContext, error) {
	dbgPrint("getNode path=%s, node=%s", path, pretty.Sprint(nc.node))
	if len(path) == 0 {
		dbgPrint("terminate with nc=%s", nc)
		return nc, nil
	}
	pe := path[0]

	v := reflect.ValueOf(nc.node)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	ncNode := v.Interface()
	// list or leaf list
	if lst, ok := ncNode.([]interface{}); ok {
		dbgPrint("list type")
		for idx, le := range lst {
			if lm, ok := le.(map[interface{}]interface{}); ok {
				k, v, err := pathKV(pe)
				if err != nil {
					return nil, err
				}
				if stringsEqual(lm[k], v) {
					dbgPrint("found matching kv %v:%v", k, v)
					nn := &nodeContext{
						parent: nc,
						node:   lm,
					}
					nc.keyToChild = idx
					nn.keyToChild = k
					if len(path) == 1 {
						dbgPrint("KV terminate")
						return nn, nil
					}
					return getNode(nn, path[1:])
				}
				continue
			}
			// Must be a leaf list
			v, err := pathV(pe)
			if err != nil {
				return nil, err
			}
			if stringsEqual(v, le) {
				dbgPrint("found matching key %v, index %d", le, idx)
				nn := &nodeContext{
					parent: nc,
					node:   le,
				}
				nc.keyToChild = idx
				return getNode(nn, path[1:])
			}
		}
		return nil, fmt.Errorf("path element %s not found", pe)
	}

	dbgPrint("interior node")
	// interior or non-leaf node
	if nnt, ok := nc.node.(map[interface{}]interface{}); ok {
		var nn interface{}
		nn = nnt[pe]
		nnc := &nodeContext{
			parent: nc,
			node:   nn,
		}
		if _, ok := nn.([]interface{}); ok {
			// Slices must be passed by pointer for mutations.
			nnc.node = &nn
		}
		nc.keyToChild = pe
		return getNode(nnc, path[1:])
	}

	return nil, fmt.Errorf("leaf type %T in non-leaf node %s", nc.node, path)
}

// writeNode writes the given value to the node in the given nodeContext.
func writeNode(nc *nodeContext, value interface{}) error {
	dbgPrint("writeNode nodeContext=%s, value=%v", nc, value)

	switch {
	case value == nil:
		dbgPrint("delete")
		switch {
		case isSlice(nc.parent.node):
			if err := util.DeleteFromSlicePtr(nc.parent.node, nc.parent.keyToChild.(int)); err != nil {
				return err
			}
			// FIXME
			if isMap(nc.parent.parent.node) {
				if err := util.InsertIntoMap(nc.parent.parent.node, nc.parent.parent.keyToChild, nc.parent.node); err != nil {
					return err
				}
			}
		}
	default:
		switch {
		case isSlice(nc.parent.node):
			idx := nc.parent.keyToChild.(int)
			if idx == -1 {
				dbgPrint("insert")

			} else {
				dbgPrint("update index %d\n", idx)
				if err := util.UpdateSlicePtr(nc.parent.node, idx, value); err != nil {
					return err
				}
			}
		default:
			dbgPrint("leaf update")
			if isMap(nc.parent.node) {
				if err := util.InsertIntoMap(nc.parent.node, nc.parent.keyToChild, value); err != nil {
					return err
				}
			}
		}
	}

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

func isSlice(v interface{}) bool {
	vv := reflect.ValueOf(v)
	if vv.Kind() == reflect.Ptr {
		vv = vv.Elem()
	}
	if vv.Kind() == reflect.Interface {
		vv = vv.Elem()
	}
	return vv.Kind() == reflect.Slice
}

func isMap(v interface{}) bool {
	vv := reflect.ValueOf(v)
	if vv.Kind() == reflect.Interface {
		vv = vv.Elem()
	}
	return vv.Kind() == reflect.Map
}

// dbgPrint prints v if the package global variable debugPackage is set.
// v has the same format as Printf. A trailing newline is added to the output.
func dbgPrint(v ...interface{}) {
	if !debugPackage {
		return
	}
	fmt.Println(fmt.Sprintf(v[0].(string), v[1:]...))
}
