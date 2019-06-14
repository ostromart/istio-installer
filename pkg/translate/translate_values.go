package translate

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/ostromart/istio-installer/pkg/component/component"
)

var (
	templateMap = map[string]*template.Template{
		component.PilotComponentName: template.Must(template.New("name").Parse(`
global:
  {{.Pilot}}
`)),
	}
)

func ValuesOverlayToValues(componentName string, componentStruct interface{}) (string, error) {
	if templateMap[componentName] == nil {
		return "", fmt.Errorf("component %s does not have a template", componentName)
	}

	var buf bytes.Buffer
	if err := templateMap[componentName].Execute(&buf, componentStruct); err != nil {
		return "", err
	}

	fmt.Println(buf.String())
	return buf.String(), nil
}
