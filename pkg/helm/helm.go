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
	YAMLSeparator = "\n---"
)

// TemplateRenderer defines a helm template renderer interface.
type TemplateRenderer interface {
	// Run starts the renderer and should be called before using it.
	Run() error
	// Render renders the associated helm charts with the given values YAML string and returns the resulting string.
	Render(values string) (string, error)
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
	if err := h.loadChart(); err != nil {
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
				if err := h.loadChart(); err != nil {
					log.Error(err.Error())
				}
			}
		}
	}()

	h.started = true
	return nil
}

func (h *FileTemplateRenderer) loadChart() error {
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

// Render renders the current helm templates with the current values and returns the resulting YAML manifest string.
func (h *FileTemplateRenderer) Render(values string) (string, error) {
	if !h.started {
		return "", fmt.Errorf("FileTemplateRenderer for %s not started in Render", h.componentName)
	}
	return Render(h.namespace, h.globalValues, values, h.chart)
}

// Render renders the given chart with the given values and returns the resulting YAML manifest string.
func Render(namespace, baseValues, overlayValues string, chrt *chart.Chart) (string, error) {
	mergedValues, err := overlayYAML(baseValues, overlayValues)
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

func overlayYAML(base, overlay string) (string, error) {
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

/*
func MergeValuesOverlay(base map[string]interface{}, overlay string) (map[string]interface{}, error) {
	bstr, err := yaml.Marshal(base)
	if err != nil {
		return nil, err
	}
	my, err := OverlayYAML(string(bstr), overlay)
	if err != nil {
		return nil, err
	}
	out := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(my), &out); err != nil {
		return nil, err
	}
	return out, nil
}
*/
