/*
Copyright 2019 The Istio Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/ghodss/yaml"
	installerv1alpha1 "github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
	"github.com/ostromart/istio-installer/pkg/controller"
	"github.com/ostromart/istio-installer/pkg/controller/istioinstaller"
	"github.com/ostromart/istio-installer/pkg/webhook"
	"github.com/ostromart/operator.old/pkg/apis"
	"istio.io/istio/pkg/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

const (
	// defaultNamespace is the default root namespace.
	defaultNamespace = "istio-system"
)

var (
	helmChartPath         = ""
	installConfigFilePath = ""
	valuesFilePath        = ""
	outputDir             = "."
	rootNamespace         = defaultNamespace
	reconcile             = false
	applyManifests        = false

	metricsAddr string
)

func init() {
	flag.StringVar(&helmChartPath, "helm-chart-path", "", "Path to helm chart directory. Can be http/git/gs or local path.")
	flag.StringVar(&installConfigFilePath, "config-file-path", "", "Path to configuration CR file (must be set if reconcile is not set).")
	flag.StringVar(&valuesFilePath, "values-file-path", "", "Path to values file (default is values.yaml in helm-chart-path).")
	flag.StringVar(&outputDir, "output-dir", ".", "Path to output directory (default is .).")
	flag.StringVar(&rootNamespace, "root-namespace", "", "Root namespace (default is "+defaultNamespace+").")
	flag.BoolVar(&reconcile, "reconcile", false, "Keep running and reconcile to any changes in charts.")
	flag.BoolVar(&applyManifests, "apply", false, "Apply the rendered manifests.")
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
}

func main() {
	flag.Parse()

	if helmChartPath == "" {
		log.Error("Must set helm-chart-path.")
		os.Exit(1)
	}
	if !reconcile && installConfigFilePath == "" {
		log.Error("Must set config-file-path if reconcile is not set.")
		os.Exit(1)
	}
	if !reconcile && valuesFilePath == "" {
		valuesFilePath = filepath.Join(helmChartPath, "values.yaml")
	}

	cleanDirPaths()

	kubeconfig, clientset, mgr, kubeclient, err := setupKube()
	if err != nil {
		log.Errora(err)
		os.Exit(1)
	}

	var baseValues map[string]interface{}
	var installConfig *installerv1alpha1.InstallerSpec
	if !reconcile {
		log.Infof("Reading install config from %s and baseValues from %s", installConfigFilePath, valuesFilePath)
		var err error
		installConfig, err = readConfigCRFromFile(installConfigFilePath)
		if err != nil {
			log.Errora(err)
			os.Exit(1)
		}
		baseValues, err = readValuesFile(valuesFilePath)
		if err != nil {
			log.Errora(err)
			os.Exit(1)
		}
	}

	dcMgr := controlplane.NewIstioInstallation(helmChartPath, kubeclient, kubeconfig, clientset, baseValues)

	switch {
	case reconcile:
		istioinstaller.SetIstioInstallationInitializer(func() *controlplane.IstioInstallation {
			return dcMgr
		})
		dcMgr.RunApplyLoop(context.Background())
		log.Info("Starting the Cmd.")
		if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
			log.Errorf("Unable to start the manager: %s", err)
			os.Exit(1)
		}

	case applyManifests:
		dcMgr.Build(installConfig)
		if err := dcMgr.ApplyOnce(context.Background()); err != nil {
			log.Errorf("ApplyOnce: %s", err)
			os.Exit(1)
		}

	default:
		dcMgr.Build(installConfig)
		if err := dcMgr.RenderToDir(outputDir); err != nil {
			log.Errorf("RenderToDir: %s", err)
			os.Exit(1)
		}
	}
}

func setupKube() (*rest.Config, *kubernetes.Clientset, manager.Manager, client.Client, error) {
	var kubeconfig *rest.Config
	var clientset *kubernetes.Clientset
	var mgr manager.Manager
	var kubeclient client.Client

	var err error

	// Don't need any kube configs since we're only rendering manifests.
	if !reconcile && !applyManifests {
		return nil, nil, nil, nil, nil
	}

	// Get a config to talk to the apiserver
	log.Info("Setting up kube config")
	kubeconfig, err = config.GetConfig()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("unable to set up client config: %s", err)
	}

	clientset, err = kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("createClientset: %s", err)
	}

	// Create a new Cmd to provide shared dependencies and start components
	log.Info("Setting up manager")
	mgr, err = manager.New(kubeconfig, manager.Options{MetricsBindAddress: metricsAddr})
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("unable to set up overall controller manager: %s", err)
	}

	// Setup Scheme for all resources
	log.Info("Setting up scheme")
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("unable add APIs to scheme: %s", err)
	}

	// Setup all Controllers
	log.Info("Setting up controller")
	if err := controller.AddToManager(mgr); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("unable to register controllers to the manager: %s", err)
	}

	log.Info("Setting up webhooks")
	if err := webhook.AddToManager(mgr); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("unable to register webhooks to the manager; %s", err)
	}

	kubeclient = mgr.GetClient()

	return kubeconfig, clientset, mgr, kubeclient, nil
}

func readConfigCRFromFile(path string) (*installerv1alpha1.InstallerSpec, error) {
	installerCR := &installerv1alpha1.IstioInstaller{}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Cannot read config file %s: %s", path, err)
	}
	if err := yaml.Unmarshal(b, installerCR); err != nil {
		return nil, fmt.Errorf("Cannot unmarshal config file %s: %s", path, err)
	}
	return installerCR.Spec, nil
}

func readValuesFile(path string) (map[string]interface{}, error) {
	values := make(map[string]interface{})
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Cannot read values file %s: %s", path, err)
	}
	if err := yaml.Unmarshal(b, &values); err != nil {
		return nil, fmt.Errorf("Cannot unmarshal config file %s: %s", path, err)
	}
	return values, nil
}

func cleanDirPaths() {
	homeDir := os.Getenv("HOME")
	helmChartPath = filepath.Clean(strings.Replace(helmChartPath, "~", homeDir, -1))
	installConfigFilePath = filepath.Clean(strings.Replace(installConfigFilePath, "~", homeDir, -1))
	outputDir = filepath.Clean(strings.Replace(outputDir, "~", homeDir, -1))
}
