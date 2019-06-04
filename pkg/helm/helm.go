package helm

import (
	"strings"

	"github.com/ostromart/istio-installer/pkg/util/fswatch"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/engine"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/timeconv"

	"istio.io/pkg/log"
)

const (
	yamlSeparator = "\n---"
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
	namespace        string
	componentName    string
	helmChartDirPath string
	watcher          chan struct{}
	chart            *chart.Chart
	values           string
}

// NewFileTemplateRenderer creates a TemplateRenderer with the given path to helm charts, k8s client config and
// ConfigSet and returns a pointer to it.
func NewFileTemplateRenderer(helmChartDirPath, componentName, namespace string) *FileTemplateRenderer {
	return &FileTemplateRenderer{
		namespace:        namespace,
		componentName:    componentName,
		helmChartDirPath: helmChartDirPath,
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

	return nil
}

func (h *FileTemplateRenderer) loadChart() error {
	var err error
	h.chart, err = chartutil.Load(h.helmChartDirPath)
	return err
}

// Render renders the current helm templates with the current values and returns the resulting YAML manifest string.
func (h *FileTemplateRenderer) Render(values string) (string, error) {
	return Render(h.namespace, values, h.chart)
}

// Render renders the given chart with the given values and returns the resulting YAML manifest string.
func Render(namespace, values string, chrt *chart.Chart) (string, error) {
	config := &chart.Config{Raw: values, Values: map[string]*chart.Value{}}
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
