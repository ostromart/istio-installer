package kube

import (
	"fmt"
	"sync"

	"github.com/ghodss/yaml"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"istio.io/istio/pkg/log"
)

// Listener is a listener for the ConfigMap with values used for helm charts.
type Listener struct {
	NotifyCh chan string

	currentValuesYAML string
	mu                sync.RWMutex

	name	   string
	namespace  string
	config     *rest.Config
	clientset  *kubernetes.Clientset
	controller cache.Controller
}

// NewListener creates a Listener object with the given k8s config and Clientset for the given namespace and returns
// a pointer to it.
func NewListener(config *rest.Config, clientset *kubernetes.Clientset, componentName, namespace string) *Listener {
	return &Listener{
		config:    config,
		clientset: clientset,
		name: componentName,
		namespace: namespace,
		NotifyCh:  make(chan string),
	}
}

func ConfigMapName(name string) string {
	return name + "-configmap"
}

// GetValues returns the YAML values string from the current ConfigMap.
func GetValues(kube kubernetes.Interface, namespace, componentName string) (string, error) {
	cm, err := kube.CoreV1().ConfigMaps(namespace).Get(ConfigMapName(componentName), metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	cmy, err := configMapToYAMLStr(cm)
	if err != nil {
		log.Error(err.Error())
	}
	return cmy, nil
}

func configMapToYAMLStr(cfgMap *v1.ConfigMap) (string, error) {
	cfg, ok := cfgMap.Data["config"]
	if !ok {
		return "", fmt.Errorf("ConfigMap yaml missing config field")
	}

	body := &configMapBody{}
	err := yaml.Unmarshal([]byte(cfg), body)
	if err != nil {
		return "", err
	}
	return body.Template, nil
}

// Listen creates a k8s Informer for the values ConfigMap and runs it in a goroutine.
func (l *Listener) Listen() {
	watchlist := cache.NewListWatchFromClient(
		l.clientset.CoreV1().RESTClient(),
		"configmaps", l.namespace,
		fields.ParseSelectorOrDie(fmt.Sprintf("metadata.name=%s", ConfigMapName(l.name))))

	_, l.controller =
		cache.NewInformer(
			watchlist,
			&v1.ConfigMap{},
			0,
			cache.ResourceEventHandlerFuncs{
				AddFunc: func(cur interface{}) {
					l.sendUpdate(cur)
				},
				UpdateFunc: func(prev, cur interface{}) {
					l.sendUpdate(cur)
				},
				DeleteFunc: func(cur interface{}) {
					log.Errorf("ConfigMap %s deleted!", ConfigMapName(l.name))
				},
			},
		)

	go func() {
		l.controller.Run(wait.NeverStop)
	}()
}

func (l *Listener) ForceUpdate() {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.NotifyCh <- l.currentValuesYAML
}

func (l *Listener) sendUpdate(configMap interface{}) {
	cm, ok := configMap.(v1.ConfigMap)
	if !ok {
		log.Errorf("sendUpdate expect ConfigMap, got %T", configMap)
		return
	}
	cmy, err := ConfigMapToYAMLStr(&cm)
	if err != nil {
		log.Error(err.Error())
	}
	l.NotifyCh <- cmy

	l.mu.Lock()
	defer l.mu.Unlock()
	l.currentValuesYAML = cmy
}

type configMapBody struct {
	Policy   string `yaml:"policy"`
	Template string `yaml:"template"`
}

func ConfigMapToYAMLStr(cfgMap *v1.ConfigMap) (string, error) {
	cfg, ok := cfgMap.Data["config"]
	if !ok {
		return "", fmt.Errorf("ConfigMap yaml missing config field")
	}

	body := &configMapBody{}
	if err := yaml.Unmarshal([]byte(cfg), body); err != nil {
		return "", err
	}

	return body.Template, nil
}
