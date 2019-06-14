package helm

import (
	"strings"

	"github.com/ostromart/istio-installer/pkg/kube"
	"github.com/ostromart/istio-installer/pkg/util/fswatch"
	"istio.io/istio/pkg/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/engine"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/timeconv"
)

const (
	yamlSeparator = "\n---"
)

// HelmTemplateRenderer is a helm template renderer.
type HelmTemplateRenderer struct {
	// ManifestCh emits new manifests whenever either the values input or the template files change.
	ManifestCh       chan string
	namespace        string
	helmChartDirPath string
	chart            *chart.Chart
	valuesListener   *kube.Listener
	curValuesYAML    string
	config           *rest.Config
	clientset        *kubernetes.Clientset
}

// NewHelmTemplateRenderer creates a HelmTemplateRenderer with the given path to helm charts, k8s client config and
// ConfigSet and returns a pointer to it.
func NewHelmTemplateRenderer(helmChartDirPath, componentName, namespace string, config *rest.Config, clientset *kubernetes.Clientset) *HelmTemplateRenderer {
	return &HelmTemplateRenderer{
		ManifestCh:       make(chan string),
		namespace:        namespace,
		helmChartDirPath: helmChartDirPath,
		valuesListener:   kube.NewListener(config, clientset, componentName, namespace),
		config:           config,
		clientset:        clientset,
	}
}

func (h *HelmTemplateRenderer) loadChart() error {
	var err error
	h.chart, err = chartutil.Load(h.helmChartDirPath)
	return err
}

// RenderNow causes a manifest to be sent on ManifestCh immediately, using the latest values and templates.
func (h *HelmTemplateRenderer) RenderNow() {
	h.valuesListener.ForceUpdate()
}

// Run creates a goroutine that listens for changes to values and template files under the configured chart directory
// and emits a rendered YAML manifest whenever either the values or templates change.
func (h *HelmTemplateRenderer) Run() error {
	var err error
	if err := h.loadChart(); err != nil {
		return err
	}

	chartChanged, err := fswatch.WatchDirRecursively(h.helmChartDirPath)
	if err != nil {
		return err
	}

	h.valuesListener.Listen()
	h.valuesListener.ForceUpdate()

	go func() {
		for {
			select {
			case <-chartChanged:
				if err := h.loadChart(); err != nil {
					log.Error(err.Error())
				}
			case h.curValuesYAML = <-h.valuesListener.NotifyCh:
			}
			mstr, err := h.Render()
			if err != nil {
				log.Error(err.Error())
				break
			}
			h.ManifestCh <- mstr
		}
	}()

	return nil
}

// RenderManifest renders the current helm templates with the current values and returns the resulting YAML manifest string.
func (h *HelmTemplateRenderer) Render() (string, error) {
	return Render(h.namespace, h.curValuesYAML, h.chart)
}

// RenderManifest renders the given chart with the given values and returns the resulting YAML manifest string.
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
		sb.WriteString(f)
	}

	return sb.String(), nil
}

/*func getHelmValues(helmChartDirectory, helmValuesFile string) string {
	valuesFile := filepath.Join(helmChartDirectory, helmValuesFile)
	return string(util.ReadFile(valuesFile, t))
}

func splitYamlFile(yamlFile string, t *testing.T) [][]byte {
	t.Helper()
	yamlBytes := util.ReadFile(yamlFile, t)
	return splitYamlBytes(yamlBytes, t)
}

func splitYamlBytes(yaml []byte, t *testing.T) [][]byte {
	t.Helper()
	stringParts := strings.Split(string(yaml), yamlSeparator)
	byteParts := make([][]byte, 0)
	for _, stringPart := range stringParts {
		byteParts = append(byteParts, getInjectableYamlDocs(stringPart, t)...)
	}
	if len(byteParts) == 0 {
		t.Skip("Found no injectable parts")
	}
	return byteParts
}

func getInjectableYamlDocs(yamlDoc string, t *testing.T) [][]byte {
	t.Helper()
	m := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(yamlDoc), &m); err != nil {
		t.Fatal(err)
	}
	return [][]byte{[]byte(yamlDoc)}
}
*/

/*fmt.Println(files)

f, ok := files[helmFilesKey]
if !ok {
	return nil, fmt.Errorf("Unable to located configmap file %s", helmFilesKey)
}

	cfgMap := core.ConfigMap{}
	err = yaml.Unmarshal([]byte(f), &cfgMap)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgMap.Data["config"]
	if !ok {
		return nil, fmt.Errorf("ConfigMap yaml missing config field")
	}

	body := &configMapBody{}
	err = yaml.Unmarshal([]byte(cfg), body)
	if err != nil {
		return nil, err
	}
*/
