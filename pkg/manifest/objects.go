package manifest

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"

	"istio.io/istio/pkg/log"
)

// Objects holds a collection of objects, so that we can filter / sequence them
type Objects struct {
	Items []*Object
}

type Object struct {
	object *unstructured.Unstructured

	Group string
	Kind  string
	Name  string
	Namespace string

	json []byte
	yaml []byte
}

func ParseJSONToObject(json []byte) (*Object, error) {
	o, gvk, err := unstructured.UnstructuredJSONScheme.Decode(json, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("error parsing json into unstructured object: %v", err)
	}

	u, ok := o.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("parsed unexpected type %T", o)
	}

	return &Object{
		object: u,
		Group:  gvk.Group,
		Kind:   gvk.Kind,
		Name:   u.GetName(),
		json:   json,
	}, nil
}

func (o *Object) AddLabels(labels map[string]string) {
	merged := make(map[string]string)
	for k, v := range o.object.GetLabels() {
		merged[k] = v
	}

	for k, v := range labels {
		merged[k] = v
	}

	o.object.SetLabels(merged)
	// Invalidate cached json
	o.json = nil
	o.yaml = nil
}

func (o *Object) NestedStringMap(fields ...string) (map[string]string, bool, error) {
	if o.object.Object == nil {
		o.object.Object = make(map[string]interface{})
	}
	return unstructured.NestedStringMap(o.object.Object, fields...)
}

func (o *Object) SetNestedField(value interface{}, fields ...string) error {
	if o.object.Object == nil {
		o.object.Object = make(map[string]interface{})
	}
	err := unstructured.SetNestedField(o.object.Object, value, fields...)
	// Invalidate cached json
	o.json = nil
	o.yaml = nil

	return err
}

func (o *Object) JSON() ([]byte, error) {
	if o.json != nil {
		return o.json, nil
	}

	b, err := o.object.MarshalJSON()
	if err != nil {
		return nil, err
	}
	o.json = b
	return b, nil
}

func (o *Object) YAML() ([]byte, error) {
	if o.yaml != nil {
		return o.yaml, nil
	}
	// TODO: there has to be a better way.
	oj, err := o.JSON()
	if err != nil {
		return nil, err
	}
	o.json = oj
	y, err := yaml.JSONToYAML(oj)
	if err != nil {
		return nil, err
	}
	o.yaml = y
	return y, nil
}

func (o *Object) YAMLDebugString() string {
	y, err := o.YAML()
	if err != nil {
		return fmt.Sprint(err)
	}
	return string(y)
}

// UnstructuredContent exposes the raw object, primarily for testing
func (o *Object) UnstructuredObject() *unstructured.Unstructured {
	return o.object
}

func (o *Object) GroupKind() schema.GroupKind {
	return o.object.GroupVersionKind().GroupKind()
}

func (o *Object) GroupVersionKind() schema.GroupVersionKind {
	return o.object.GroupVersionKind()
}

func (o *Object) Hash() string {
	return Hash(o.Kind, o.Namespace, o.Name)
}

func (o *Objects) JSONManifest() (string, error) {
	var b bytes.Buffer

	for i, item := range o.Items {
		if i != 0 {
			b.WriteString("\n\n")
		}
		// We build a JSON manifest because conversion to yaml is harder
		// (and we've lost line numbers anyway if we applied any transforms)
		json, err := item.JSON()
		if err != nil {
			return "", fmt.Errorf("error building json: %v", err)
		}
		b.Write(json)
	}

	return b.String(), nil
}

// Sort will order the items in Objects in order of score, group, kind, name.  The intent is to
// have a deterministic ordering in which Objects are applied.
func (o *Objects) Sort(score func(o *Object) int) {
	sort.Slice(o.Items, func(i, j int) bool {
		iScore := score(o.Items[i])
		jScore := score(o.Items[j])
		return iScore < jScore ||
			(iScore == jScore &&
				o.Items[i].Group < o.Items[j].Group) ||
			(iScore == jScore &&
				o.Items[i].Group == o.Items[j].Group &&
				o.Items[i].Kind < o.Items[j].Kind) ||
			(iScore == jScore &&
				o.Items[i].Group == o.Items[j].Group &&
				o.Items[i].Kind == o.Items[j].Kind &&
				o.Items[i].Name < o.Items[j].Name)
	})
}

func (o *Objects) ToMap() map[string]*Object {
	ret := make(map[string]*Object)
	for _, oo := range o.Items {
		ret[oo.Hash()] = oo
	}
	return ret
}

func ParseObjectsFromYAMLManifest(ctx context.Context, manifest string) (*Objects, error) {
	var b bytes.Buffer

	var yamls []string
	for _, line := range strings.Split(manifest, "\n") {
		if line == "---" {
			// yaml separator
			yamls = append(yamls, b.String())
			b.Reset()
		} else {
			b.WriteString(line)
			b.WriteString("\n")
		}
	}
	yamls = append(yamls, b.String())

	objects := &Objects{}

	for _, yaml := range yamls {
		// We need this so we don't error on a file that is commented out
		// TODO: How does apimachinery avoid this problem?
		hasContent := false
		for _, line := range strings.Split(yaml, "\n") {
			l := strings.TrimSpace(line)
			if l != "" && !strings.HasPrefix(l, "#") {
				hasContent = true
				break
			}
		}

		if !hasContent {
			continue
		}

		r := bytes.NewReader([]byte(yaml))
		decoder := k8syaml.NewYAMLOrJSONDecoder(r, 1024)

		out := &unstructured.Unstructured{}
		err := decoder.Decode(out)
		if err != nil {
			log.Infof("error decoding object: %s\n%s\n", err, yaml)
			return nil, fmt.Errorf("error decoding object: %v", err)
		}

		var json []byte
		// We don't reuse the manifest because it's probably yaml, and we want to use json
		// json = yaml
		o := NewObject(out, json, []byte(yaml))
		objects.Items = append(objects.Items, o)
	}

	return objects, nil
}

func ObjectsFromUnstructuredSlice(objs []*unstructured.Unstructured) (*Objects, error) {
	ret := &Objects{}
	for _, o := range objs {
		ret.Items = append(ret.Items, NewObject(o, nil, nil))
	}
	return ret, nil
}

func NewObject(u *unstructured.Unstructured, json, yaml []byte) *Object {
	o := &Object{
		object: u,
		json:   json,
		yaml: yaml,
	}

	gvk := u.GetObjectKind().GroupVersionKind()
	o.Group = gvk.Group
	o.Kind = gvk.Kind
	o.Name = u.GetName()
	o.Namespace = u.GetNamespace()

	return o
}

func Hash(kind, namespace, name string) string {
	return strings.Join([]string{kind, namespace, name}, "/")
}