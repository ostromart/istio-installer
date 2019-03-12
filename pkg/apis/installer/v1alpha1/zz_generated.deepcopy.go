// +build !ignore_autogenerated

/*
Copyright The Kubernetes Authors.

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

// Code generated by main. DO NOT EDIT.

package v1alpha1

import (
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/policy/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConfigManagementConfig) DeepCopyInto(out *ConfigManagementConfig) {
	*out = *in
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConfigManagementConfig.
func (in *ConfigManagementConfig) DeepCopy() *ConfigManagementConfig {
	if in == nil {
		return nil
	}
	out := new(ConfigManagementConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EgressGatewayConfig) DeepCopyInto(out *EgressGatewayConfig) {
	*out = *in
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EgressGatewayConfig.
func (in *EgressGatewayConfig) DeepCopy() *EgressGatewayConfig {
	if in == nil {
		return nil
	}
	out := new(EgressGatewayConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IngressGatewayConfig) DeepCopyInto(out *IngressGatewayConfig) {
	*out = *in
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IngressGatewayConfig.
func (in *IngressGatewayConfig) DeepCopy() *IngressGatewayConfig {
	if in == nil {
		return nil
	}
	out := new(IngressGatewayConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InstallerSpec) DeepCopyInto(out *InstallerSpec) {
	*out = *in
	if in.InstallPackagePath != nil {
		in, out := &in.InstallPackagePath, &out.InstallPackagePath
		*out = new(string)
		**out = **in
	}
	if in.ControllerName != nil {
		in, out := &in.ControllerName, &out.ControllerName
		*out = new(string)
		**out = **in
	}
	if in.Hub != nil {
		in, out := &in.Hub, &out.Hub
		*out = new(string)
		**out = **in
	}
	if in.Tag != nil {
		in, out := &in.Tag, &out.Tag
		*out = new(string)
		**out = **in
	}
	if in.DefaultNamespace != nil {
		in, out := &in.DefaultNamespace, &out.DefaultNamespace
		*out = new(string)
		**out = **in
	}
	if in.RemoteClusterConfig != nil {
		in, out := &in.RemoteClusterConfig, &out.RemoteClusterConfig
		*out = new(RemoteClusterConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.TrafficManagement != nil {
		in, out := &in.TrafficManagement, &out.TrafficManagement
		*out = new(TrafficManagementConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.Policy != nil {
		in, out := &in.Policy, &out.Policy
		*out = new(PolicyConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.Telemetry != nil {
		in, out := &in.Telemetry, &out.Telemetry
		*out = new(TelemetryConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.Security != nil {
		in, out := &in.Security, &out.Security
		*out = new(SecurityConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.ConfigManagement != nil {
		in, out := &in.ConfigManagement, &out.ConfigManagement
		*out = new(ConfigManagementConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.IngressGateway != nil {
		in, out := &in.IngressGateway, &out.IngressGateway
		*out = make([]*IngressGatewayConfig, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(IngressGatewayConfig)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.EgressGateway != nil {
		in, out := &in.EgressGateway, &out.EgressGateway
		*out = make([]*EgressGatewayConfig, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(EgressGatewayConfig)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.ExternalOperators != nil {
		in, out := &in.ExternalOperators, &out.ExternalOperators
		*out = make([]*OperatorConfig, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(OperatorConfig)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InstallerSpec.
func (in *InstallerSpec) DeepCopy() *InstallerSpec {
	if in == nil {
		return nil
	}
	out := new(InstallerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InstallerStatus) DeepCopyInto(out *InstallerStatus) {
	*out = *in
	if in.ProxyControl != nil {
		in, out := &in.ProxyControl, &out.ProxyControl
		*out = new(string)
		**out = **in
	}
	if in.SidecarInjector != nil {
		in, out := &in.SidecarInjector, &out.SidecarInjector
		*out = new(string)
		**out = **in
	}
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InstallerStatus.
func (in *InstallerStatus) DeepCopy() *InstallerStatus {
	if in == nil {
		return nil
	}
	out := new(InstallerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IstioInstaller) DeepCopyInto(out *IstioInstaller) {
	*out = *in
	if in.Spec != nil {
		in, out := &in.Spec, &out.Spec
		*out = new(InstallerSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Status != nil {
		in, out := &in.Status, &out.Status
		*out = new(InstallerStatus)
		(*in).DeepCopyInto(*out)
	}
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IstioInstaller.
func (in *IstioInstaller) DeepCopy() *IstioInstaller {
	if in == nil {
		return nil
	}
	out := new(IstioInstaller)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Object) DeepCopyInto(out *Object) {
	*out = *in
	if in.Metadata != nil {
		in, out := &in.Metadata, &out.Metadata
		*out = new(v1.ObjectMeta)
		(*in).DeepCopyInto(*out)
	}
	if in.Data != nil {
		in, out := &in.Data, &out.Data
		*out = new(runtime.RawExtension)
		(*in).DeepCopyInto(*out)
	}
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Object.
func (in *Object) DeepCopy() *Object {
	if in == nil {
		return nil
	}
	out := new(Object)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OperatorConfig) DeepCopyInto(out *OperatorConfig) {
	*out = *in
	if in.ManifestPath != nil {
		in, out := &in.ManifestPath, &out.ManifestPath
		*out = new(string)
		**out = **in
	}
	if in.Namespace != nil {
		in, out := &in.Namespace, &out.Namespace
		*out = new(string)
		**out = **in
	}
	if in.Spec != nil {
		in, out := &in.Spec, &out.Spec
		*out = new(Object)
		(*in).DeepCopyInto(*out)
	}
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OperatorConfig.
func (in *OperatorConfig) DeepCopy() *OperatorConfig {
	if in == nil {
		return nil
	}
	out := new(OperatorConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PilotConfig) DeepCopyInto(out *PilotConfig) {
	*out = *in
	if in.Debug != nil {
		in, out := &in.Debug, &out.Debug
		*out = new(bool)
		**out = **in
	}
	if in.Sidecar != nil {
		in, out := &in.Sidecar, &out.Sidecar
		*out = new(bool)
		**out = **in
	}
	if in.TraceSampling != nil {
		in, out := &in.TraceSampling, &out.TraceSampling
		*out = new(float32)
		**out = **in
	}
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(corev1.ResourceRequirements)
		(*in).DeepCopyInto(*out)
	}
	if in.HpaSpec != nil {
		in, out := &in.HpaSpec, &out.HpaSpec
		*out = new(autoscalingv1.HorizontalPodAutoscalerSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.PodDisruptionBudget != nil {
		in, out := &in.PodDisruptionBudget, &out.PodDisruptionBudget
		*out = new(v1beta1.PodDisruptionBudget)
		(*in).DeepCopyInto(*out)
	}
	if in.NodeSelector != nil {
		in, out := &in.NodeSelector, &out.NodeSelector
		*out = new(corev1.NodeSelector)
		(*in).DeepCopyInto(*out)
	}
	if in.AdditionalArgs != nil {
		in, out := &in.AdditionalArgs, &out.AdditionalArgs
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.ResourceOverride != nil {
		in, out := &in.ResourceOverride, &out.ResourceOverride
		*out = make([]*Object, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(Object)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PilotConfig.
func (in *PilotConfig) DeepCopy() *PilotConfig {
	if in == nil {
		return nil
	}
	out := new(PilotConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolicyConfig) DeepCopyInto(out *PolicyConfig) {
	*out = *in
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolicyConfig.
func (in *PolicyConfig) DeepCopy() *PolicyConfig {
	if in == nil {
		return nil
	}
	out := new(PolicyConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProxyConfig) DeepCopyInto(out *ProxyConfig) {
	*out = *in
	if in.Debug != nil {
		in, out := &in.Debug, &out.Debug
		*out = new(bool)
		**out = **in
	}
	if in.Privileged != nil {
		in, out := &in.Privileged, &out.Privileged
		*out = new(bool)
		**out = **in
	}
	if in.EnableCoredump != nil {
		in, out := &in.EnableCoredump, &out.EnableCoredump
		*out = new(bool)
		**out = **in
	}
	if in.InterceptionMode != nil {
		in, out := &in.InterceptionMode, &out.InterceptionMode
		*out = new(ProxyConfig_InterceptionMode)
		**out = **in
	}
	if in.StatusPort != nil {
		in, out := &in.StatusPort, &out.StatusPort
		*out = new(uint32)
		**out = **in
	}
	if in.ImagePullPolicy != nil {
		in, out := &in.ImagePullPolicy, &out.ImagePullPolicy
		*out = new(string)
		**out = **in
	}
	if in.ProxyInitImage != nil {
		in, out := &in.ProxyInitImage, &out.ProxyInitImage
		*out = new(string)
		**out = **in
	}
	if in.IncludeIpRanges != nil {
		in, out := &in.IncludeIpRanges, &out.IncludeIpRanges
		*out = new(string)
		**out = **in
	}
	if in.ExcludeIpRanges != nil {
		in, out := &in.ExcludeIpRanges, &out.ExcludeIpRanges
		*out = new(string)
		**out = **in
	}
	if in.IncludeInboundPorts != nil {
		in, out := &in.IncludeInboundPorts, &out.IncludeInboundPorts
		*out = new(string)
		**out = **in
	}
	if in.ExcludeInboundPorts != nil {
		in, out := &in.ExcludeInboundPorts, &out.ExcludeInboundPorts
		*out = new(string)
		**out = **in
	}
	if in.ConnectTimeout != nil {
		in, out := &in.ConnectTimeout, &out.ConnectTimeout
		*out = new(string)
		**out = **in
	}
	if in.DrainDuration != nil {
		in, out := &in.DrainDuration, &out.DrainDuration
		*out = new(string)
		**out = **in
	}
	if in.ParentShutdownDuration != nil {
		in, out := &in.ParentShutdownDuration, &out.ParentShutdownDuration
		*out = new(string)
		**out = **in
	}
	if in.Concurrency != nil {
		in, out := &in.Concurrency, &out.Concurrency
		*out = new(uint32)
		**out = **in
	}
	if in.ClusterDomain != nil {
		in, out := &in.ClusterDomain, &out.ClusterDomain
		*out = new(string)
		**out = **in
	}
	if in.PodDnsSearchNamespaces != nil {
		in, out := &in.PodDnsSearchNamespaces, &out.PodDnsSearchNamespaces
		*out = new(string)
		**out = **in
	}
	if in.Lightstep != nil {
		in, out := &in.Lightstep, &out.Lightstep
		*out = new(ProxyConfig_LightstepConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.Zipkin != nil {
		in, out := &in.Zipkin, &out.Zipkin
		*out = new(ProxyConfig_ZipkinConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.Sds != nil {
		in, out := &in.Sds, &out.Sds
		*out = new(SdsConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.ReadinessProbe != nil {
		in, out := &in.ReadinessProbe, &out.ReadinessProbe
		*out = new(corev1.Probe)
		(*in).DeepCopyInto(*out)
	}
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(corev1.ResourceRequirements)
		(*in).DeepCopyInto(*out)
	}
	if in.AdditionalArgs != nil {
		in, out := &in.AdditionalArgs, &out.AdditionalArgs
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.ResourceOverride != nil {
		in, out := &in.ResourceOverride, &out.ResourceOverride
		*out = make([]*Object, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(Object)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProxyConfig.
func (in *ProxyConfig) DeepCopy() *ProxyConfig {
	if in == nil {
		return nil
	}
	out := new(ProxyConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProxyConfig_LightstepConfig) DeepCopyInto(out *ProxyConfig_LightstepConfig) {
	*out = *in
	if in.Address != nil {
		in, out := &in.Address, &out.Address
		*out = new(string)
		**out = **in
	}
	if in.AccessToken != nil {
		in, out := &in.AccessToken, &out.AccessToken
		*out = new(string)
		**out = **in
	}
	if in.CaCertPath != nil {
		in, out := &in.CaCertPath, &out.CaCertPath
		*out = new(string)
		**out = **in
	}
	if in.Secure != nil {
		in, out := &in.Secure, &out.Secure
		*out = new(bool)
		**out = **in
	}
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProxyConfig_LightstepConfig.
func (in *ProxyConfig_LightstepConfig) DeepCopy() *ProxyConfig_LightstepConfig {
	if in == nil {
		return nil
	}
	out := new(ProxyConfig_LightstepConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProxyConfig_ZipkinConfig) DeepCopyInto(out *ProxyConfig_ZipkinConfig) {
	*out = *in
	if in.Address != nil {
		in, out := &in.Address, &out.Address
		*out = new(string)
		**out = **in
	}
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProxyConfig_ZipkinConfig.
func (in *ProxyConfig_ZipkinConfig) DeepCopy() *ProxyConfig_ZipkinConfig {
	if in == nil {
		return nil
	}
	out := new(ProxyConfig_ZipkinConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RemoteClusterConfig) DeepCopyInto(out *RemoteClusterConfig) {
	*out = *in
	if in.Enabled != nil {
		in, out := &in.Enabled, &out.Enabled
		*out = new(bool)
		**out = **in
	}
	if in.Namespace != nil {
		in, out := &in.Namespace, &out.Namespace
		*out = new(string)
		**out = **in
	}
	if in.ResourceOverride != nil {
		in, out := &in.ResourceOverride, &out.ResourceOverride
		*out = make([]*Object, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(Object)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RemoteClusterConfig.
func (in *RemoteClusterConfig) DeepCopy() *RemoteClusterConfig {
	if in == nil {
		return nil
	}
	out := new(RemoteClusterConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SdsConfig) DeepCopyInto(out *SdsConfig) {
	*out = *in
	if in.Enabled != nil {
		in, out := &in.Enabled, &out.Enabled
		*out = new(bool)
		**out = **in
	}
	if in.UdsPath != nil {
		in, out := &in.UdsPath, &out.UdsPath
		*out = new(string)
		**out = **in
	}
	if in.UseTrustworthyJwt != nil {
		in, out := &in.UseTrustworthyJwt, &out.UseTrustworthyJwt
		*out = new(bool)
		**out = **in
	}
	if in.UseNormalJwt != nil {
		in, out := &in.UseNormalJwt, &out.UseNormalJwt
		*out = new(bool)
		**out = **in
	}
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SdsConfig.
func (in *SdsConfig) DeepCopy() *SdsConfig {
	if in == nil {
		return nil
	}
	out := new(SdsConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecurityConfig) DeepCopyInto(out *SecurityConfig) {
	*out = *in
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecurityConfig.
func (in *SecurityConfig) DeepCopy() *SecurityConfig {
	if in == nil {
		return nil
	}
	out := new(SecurityConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TelemetryConfig) DeepCopyInto(out *TelemetryConfig) {
	*out = *in
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TelemetryConfig.
func (in *TelemetryConfig) DeepCopy() *TelemetryConfig {
	if in == nil {
		return nil
	}
	out := new(TelemetryConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TrafficManagementConfig) DeepCopyInto(out *TrafficManagementConfig) {
	*out = *in
	if in.Enabled != nil {
		in, out := &in.Enabled, &out.Enabled
		*out = new(bool)
		**out = **in
	}
	if in.Namespace != nil {
		in, out := &in.Namespace, &out.Namespace
		*out = new(string)
		**out = **in
	}
	if in.AutoInjection != nil {
		in, out := &in.AutoInjection, &out.AutoInjection
		*out = new(bool)
		**out = **in
	}
	if in.PilotConfig != nil {
		in, out := &in.PilotConfig, &out.PilotConfig
		*out = new(PilotConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.ProxyConfig != nil {
		in, out := &in.ProxyConfig, &out.ProxyConfig
		*out = new(ProxyConfig)
		(*in).DeepCopyInto(*out)
	}
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TrafficManagementConfig.
func (in *TrafficManagementConfig) DeepCopy() *TrafficManagementConfig {
	if in == nil {
		return nil
	}
	out := new(TrafficManagementConfig)
	in.DeepCopyInto(out)
	return out
}
