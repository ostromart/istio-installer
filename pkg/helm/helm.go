package helm

import (
	"fmt"
	"io/ioutil"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/ghodss/yaml"

	"github.com/ostromart/istio-installer/pkg/util/fswatch"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/engine"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/timeconv"

	"istio.io/pkg/log"
)

const (
	// YAMLSeparator is a separator for multi-document YAML files.
	YAMLSeparator = "\n---"

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
	// LoadChart loads the chart from the associated chart source.
	LoadChart() error
}

// FileTemplateRenderer is a helm template renderer.
type FileTemplateRenderer struct {
	namespace            string
	componentName        string
	globalValuesFilePath string
	helmChartDirPath     string
	watcher              chan struct{}
	chart                *chart.Chart
	values               string
	started              bool
	globalValues         string
}

// NewFileTemplateRenderer creates a TemplateRenderer with the given path to helm charts, k8s client config and
// ConfigSet and returns a pointer to it.
func NewFileTemplateRenderer(globalValuesFilePath, helmChartDirPath, componentName, namespace string) *FileTemplateRenderer {
	fmt.Printf("NewFileTemplateRenderer with helmChart=%s, globalVals=%s\n", helmChartDirPath, globalValuesFilePath)
	return &FileTemplateRenderer{
		namespace:            namespace,
		componentName:        componentName,
		globalValuesFilePath: globalValuesFilePath,
		helmChartDirPath:     helmChartDirPath,
	}
}

// Run implements the TemplateRenderer interface.
func (h *FileTemplateRenderer) Run() error {
	var err error
	if err := h.LoadChart(); err != nil {
		return err
	}

	chartChanged, err := fswatch.WatchDirRecursively(h.helmChartDirPath)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-chartChanged:
				if err := h.LoadChart(); err != nil {
					log.Error(err.Error())
				}
			}
		}
	}()

	h.started = true
	return nil
}

// LoadChart implements the TemplateRenderer interface.
func (h *FileTemplateRenderer) LoadChart() error {
	var err error
	if h.chart, err = chartutil.Load(h.helmChartDirPath); err != nil {
		return err
	}
	b, err := ioutil.ReadFile(h.globalValuesFilePath)
	if err != nil {
		return err
	}
	h.globalValues = string(b)
	return nil
}

// RenderManifest renders the current helm templates with the current values and returns the resulting YAML manifest string.
func (h *FileTemplateRenderer) RenderManifest(values string) (string, error) {
	if !h.started {
		return "", fmt.Errorf("FileTemplateRenderer for %s not started in renderChart", h.componentName)
	}
	return renderChart(h.namespace, h.globalValues, values, h.chart)
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
		return "", fmt.Errorf("YAMLToJSON error in base: %s\n%s\n", err, bj)
	}
	oj, err := yaml.YAMLToJSON([]byte(overlay))
	if err != nil {
		return "", fmt.Errorf("YAMLToJSON error in overlay: %s\n%s\n", err, oj)
	}

	merged, err := jsonpatch.MergePatch(bj, oj)
	if err != nil {
		return "", fmt.Errorf("JSON merge error (%s) for base object: \n%s\n override object: \n%s", err, bj, oj)
	}
	my, err := yaml.JSONToYAML(merged)
	if err != nil {
		return "", fmt.Errorf("JSONToYAML error (%s) for merged object: \n%s", err, merged)
	}

	return string(my), nil
}
