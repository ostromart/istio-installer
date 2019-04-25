package patch

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/evanphx/json-patch"
	"github.com/ghodss/yaml"
	"github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
	"github.com/ostromart/istio-installer/pkg/manifest"
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
