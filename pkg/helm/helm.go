// Copyright 2019 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helm

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/ostromart/istio-installer/pkg/util"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/ghodss/yaml"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/engine"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/timeconv"
)

const (
	// YAMLSeparator is a separator for multi-document YAML files.
	YAMLSeparator = "\n---\n"

	// DefaultGlobalValuesFilename is the default name for a global values file if none is specified.
	DefaultGlobalValuesFilename = "global.yaml"
)

// TemplateRenderer defines a helm template renderer interface.
type TemplateRenderer interface {
	// Run starts the renderer and should be called before using it.
	Run() error
	// RenderManifest renders the associated helm charts with the given values YAML string and returns the resulting
	// string.
	RenderManifest(values string) (string, error)
}

// NewHelmRenderer creates a new helm renderer with the given parameters and returns an interface to it.
// The format of helmBaseDir and profile strings determines the type of helm renderer returned (compiled-in, file,
// HTTP etc.)
func NewHelmRenderer(helmBaseDir, profile, componentName, namespace string) (TemplateRenderer, error) {
	globalValues, err := ReadValuesYAML(profile)
	if err != nil {
		return nil, err
	}
	switch {
	case util.IsFilePath(helmBaseDir):
		return NewFileTemplateRenderer(helmBaseDir, globalValues, componentName, namespace), nil
	default:
		return NewVFSRenderer(helmBaseDir, globalValues, componentName, namespace), nil
	}
	return nil, fmt.Errorf("unsupported CustomPackagePath: %s", helmBaseDir)
}

// ReadValuesYAML reads the values YAML associated with the given profile. It uses an appropriate reader for the
// profile format (compiled-in, file, HTTP, etc.).
func ReadValuesYAML(profile string) (string, error) {
	var err error
	var globalValues string

	// Get global values from profile.
	switch {
	case isBuiltinProfileName(profile):
		if globalValues, err = LoadValuesVFS(profile); err != nil {
			return "", err
		}
	case util.IsFilePath(profile):
		if globalValues, err = readFile(util.GetLocalFilePath(profile)); err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unsupported Profile type: %s", profile)
	}

	return globalValues, nil
}

// renderChart renders the given chart with the given values and returns the resulting YAML manifest string.
func renderChart(namespace, baseValues, overlayValues string, chrt *chart.Chart) (string, error) {
	mergedValues, err := OverlayYAML(baseValues, overlayValues)
	if err != nil {
		return "", err
	}

	config := &chart.Config{Raw: mergedValues, Values: map[string]*chart.Value{}}
	options := chartutil.ReleaseOptions{
		Name:      "istio",
		Time:      timeconv.Now(),
		Namespace: namespace,
	}

	vals, err := chartutil.ToRenderValuesCaps(chrt, config, options, nil)
	if err != nil {
		return "", err
	}

	files, err := engine.New().Render(chrt, vals)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, f := range files {
		_, err := sb.WriteString(f)
		if err != nil {
			return "", err
		}
	}

	return sb.String(), nil
}

// OverlayYAML patches the overlay tree over the base tree and returns the result. All trees are expressed as YAML
// strings.
func OverlayYAML(base, overlay string) (string, error) {
	bj, err := yaml.YAMLToJSON([]byte(base))
	if err != nil {
		return "", fmt.Errorf("yAMLToJSON error in base: %s\n%s\n", err, bj)
	}
	oj, err := yaml.YAMLToJSON([]byte(overlay))
	if err != nil {
		return "", fmt.Errorf("yAMLToJSON error in overlay: %s\n%s\n", err, oj)
	}

	fmt.Printf("HERE: base:\n%s\n\noverlay:\n%s\n", base, overlay)

	merged, err := jsonpatch.MergePatch(bj, oj)
	if err != nil {
		return "", fmt.Errorf("jSON merge error (%s) for base object: \n%s\n override object: \n%s", err, bj, oj)
	}
	my, err := yaml.JSONToYAML(merged)
	if err != nil {
		return "", fmt.Errorf("jSONToYAML error (%s) for merged object: \n%s", err, merged)
	}

	return string(my), nil
}

func readFile(path string) (string, error) {
	b, err := ioutil.ReadFile(path)
	return string(b), err
}
