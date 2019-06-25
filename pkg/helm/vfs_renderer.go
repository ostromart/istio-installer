package helm

import (
	"fmt"
	"path/filepath"
	"strings"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/ostromart/istio-installer/pkg/vfsgen"

	"istio.io/pkg/log"
)

const (
	// DefaultProfileFilename is the name of the default profile yaml file.
	DefaultProfileFilename = "default.yaml"

	chartsRoot   = "/charts"
	profilesRoot = "/profiles"
)

var (
	// ProfileNames holds the names of all the profiles in the /profiles directory, without .yaml suffix.
	ProfileNames = make(map[string]bool)
)

func init() {
	profilePaths, err := vfsgen.ReadDir(profilesRoot)
	if err != nil {
		panic(err)
	}
	for _, p := range profilePaths {
		p = strings.TrimSuffix(p, ".yaml")
		ProfileNames[p] = true
	}
}

// VFSRenderer is a helm template renderer that uses compiled-in helm charts.
type VFSRenderer struct {
	namespace        string
	componentName    string
	helmChartDirPath string
	chart            *chart.Chart
	started          bool
	valuesYAML       string
}

// NewVFSRenderer creates a VFSRenderer with the given relative path to helm charts, component name and namespace and
// a base values YAML string.
func NewVFSRenderer(helmChartDirPath, valuesYAML, componentName, namespace string) *VFSRenderer {
	log.Infof("NewVFSRenderer with helmChart=%s, componentName=%s", helmChartDirPath, componentName)
	return &VFSRenderer{
		namespace:        namespace,
		componentName:    componentName,
		helmChartDirPath: helmChartDirPath,
		valuesYAML:       valuesYAML,
	}
}

// Run implements the TemplateRenderer interface.
func (h *VFSRenderer) Run() error {
	log.Infof("Run FileTemplateRenderer with helmChart=%s, componentName=%s", h.helmChartDirPath, h.componentName)
	if err := h.loadChart(); err != nil {
		return err
	}
	h.started = true
	return nil
}

// RenderManifest renders the current helm templates with the current values and returns the resulting YAML manifest
// string.
func (h *VFSRenderer) RenderManifest(values string) (string, error) {
	if !h.started {
		return "", fmt.Errorf("VFSRenderer for %s not started in renderChart", h.componentName)
	}
	return renderChart(h.namespace, h.valuesYAML, values, h.chart)
}

// LoadValuesVFS loads the compiled in file corresponding to the given profile name.
func LoadValuesVFS(profileName string) (string, error) {
	path := filepath.Join(profilesRoot, BuiltinProfileToFilename(profileName))
	log.Infof("Loading values from compiled in VFS at path %s", path)
	b, err := vfsgen.ReadFile(path)
	return string(b), err
}

func isBuiltinProfileName(name string) bool {
	if name == "" {
		return true
	}
	return ProfileNames[name]
}

// loadChart implements the TemplateRenderer interface.
func (h *VFSRenderer) loadChart() error {
	prefix := filepath.Join(chartsRoot, h.helmChartDirPath)
	fnames, err := vfsgen.GetFilesRecursive(prefix)
	if err != nil {
		return err
	}
	var bfs []*chartutil.BufferedFile
	for _, fname := range fnames {
		b, err := vfsgen.ReadFile(fname)
		if err != nil {
			return err
		}
		bf := &chartutil.BufferedFile{
			Name: stripPrefix(fname, prefix),
			Data: b,
		}
		bfs = append(bfs, bf)
		log.Infof("loaded %s", bf.Name)
	}

	h.chart, err = chartutil.LoadFiles(bfs)
	return err
}

func BuiltinProfileToFilename(name string) string {
	if name == "" {
		return DefaultProfileFilename
	}
	return name + ".yaml"
}

// stripPrefix removes the the given prefix from prefix.
func stripPrefix(path, prefix string) string {
	pl := len(strings.Split(prefix, "/"))
	pv := strings.Split(path, "/")
	return strings.Join(pv[pl:], "/")
}
