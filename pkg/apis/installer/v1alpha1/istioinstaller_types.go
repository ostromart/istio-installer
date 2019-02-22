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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ComponentOptions struct {
	Enabled   bool   `json:"enabled,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// IstioInstallerSpec defines the desired state of IstioInstaller
type IstioInstallerSpec struct {
	ChartPath     string `json:"chartPath,omitempty"`
	Version       string `json:"version,omitempty"`
	RootNamespace string `json:"namespace,omitempty"`

	InstallProxyControl     ComponentOptions `json:"installProxyControl,omitempty"`
	InstallSidecarInjection ComponentOptions `json:"installSidecarInjection,omitempty"`
	InstallIngress          ComponentOptions `json:"installIngress,omitempty"`
	InstallIngressGateway   ComponentOptions `json:"installIngressGateway,omitempty"`
	InstallEgressGateway    ComponentOptions `json:"installEgressGateway,omitempty"`
	InstallPolicy           ComponentOptions `json:"installPolicy,omitempty"`
	InstallTelemetry        ComponentOptions `json:"installTelemetry,omitempty"`
	InstallSecurity         ComponentOptions `json:"installSecurity,omitempty"`
	InstallConfigManagement ComponentOptions `json:"installConfigManagement,omitempty"`
	InstallCoreDNS          ComponentOptions `json:"installCoreDNS,omitempty"`
	InstallCNI              ComponentOptions `json:"installCNI,omitempty"`

	InstallGrafana      ComponentOptions `json:"installGrafana,omitempty"`
	InstallPrometheus   ComponentOptions `json:"installPrometheus,omitempty"`
	InstallKiali        ComponentOptions `json:"installKiali,omitempty"`
	InstallServiceGraph ComponentOptions `json:"installServiceGraph,omitempty"`
	InstallTracing      ComponentOptions `json:"installTracing,omitempty"`

	ValuesOverride   interface{}                  `json:"valuesOverride,omitempty"`
	ResourceOverride []*unstructured.Unstructured `json:"resourceOverride,omitempty"`
}

// IstioInstallerStatus defines the observed state of IstioInstaller
type IstioInstallerStatus struct {
	ProxyControlStatus     string `json:"proxyControlStatus,omitempty"`
	SidecarInjectionStatus string `json:"sidecarInjectionStatus,omitempty"`
	IngressStatus          string `json:"ingressStatus,omitempty"`
	IngressGatewayStatus   string `json:"ingressGatewayStatus,omitempty"`
	EgressGatewayStatus    string `json:"egressGatewayStatus,omitempty"`
	PolicyStatus           string `json:"policyStatus,omitempty"`
	TelemetryStatus        string `json:"telemetryStatus,omitempty"`
	SecurityStatus         string `json:"securityStatus,omitempty"`
	ConfigManagementStatus string `json:"configManagementStatus,omitempty"`
	CoreDNSStatus          string `json:"coreDNSStatus,omitempty"`
	CNIStatus              string `json:"cNIStatus,omitempty"`

	GrafanaStatus      string `json:"grafanaStatus,omitempty"`
	PrometheusStatus   string `json:"prometheusStatus,omitempty"`
	KialiStatus        string `json:"kialiStatus,omitempty"`
	ServiceGraphStatus string `json:"serviceGraphStatus,omitempty"`
	TracingStatus      string `json:"tracingStatus,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IstioInstaller is the Schema for the istioinstallers API
// +k8s:openapi-gen=true
type IstioInstaller struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IstioInstallerSpec   `json:"spec,omitempty"`
	Status IstioInstallerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IstioInstallerList contains a list of IstioInstaller
type IstioInstallerList struct {
	metav1.TypeMeta        `json:",inline"`
	metav1.ListMeta        `json:"metadata,omitempty"`
	Items []IstioInstaller `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IstioInstaller{}, &IstioInstallerList{})
}
