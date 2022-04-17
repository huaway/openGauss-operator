//go:build !ignore_autogenerated
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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1

import (
	corev1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenGauss) DeepCopyInto(out *OpenGauss) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	if in.Status != nil {
		in, out := &in.Status, &out.Status
		*out = new(OpenGaussStatus)
		(*in).DeepCopyInto(*out)
	}
	if in.Spec != nil {
		in, out := &in.Spec, &out.Spec
		*out = new(OpenGaussSpec)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenGauss.
func (in *OpenGauss) DeepCopy() *OpenGauss {
	if in == nil {
		return nil
	}
	out := new(OpenGauss)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OpenGauss) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenGaussClusterConfiguration) DeepCopyInto(out *OpenGaussClusterConfiguration) {
	*out = *in
	if in.Master != nil {
		in, out := &in.Master, &out.Master
		*out = new(OpenGaussStatefulSet)
		(*in).DeepCopyInto(*out)
	}
	if in.WorkerSmall != nil {
		in, out := &in.WorkerSmall, &out.WorkerSmall
		*out = new(OpenGaussStatefulSet)
		(*in).DeepCopyInto(*out)
	}
	if in.WorkerMid != nil {
		in, out := &in.WorkerMid, &out.WorkerMid
		*out = new(OpenGaussStatefulSet)
		(*in).DeepCopyInto(*out)
	}
	if in.WorkerLarge != nil {
		in, out := &in.WorkerLarge, &out.WorkerLarge
		*out = new(OpenGaussStatefulSet)
		(*in).DeepCopyInto(*out)
	}
	if in.Shardingsphere != nil {
		in, out := &in.Shardingsphere, &out.Shardingsphere
		*out = new(ShardingsphereStatefulSet)
		(*in).DeepCopyInto(*out)
	}
	if in.Origin != nil {
		in, out := &in.Origin, &out.Origin
		*out = new(OriginOpenGaussCluster)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenGaussClusterConfiguration.
func (in *OpenGaussClusterConfiguration) DeepCopy() *OpenGaussClusterConfiguration {
	if in == nil {
		return nil
	}
	out := new(OpenGaussClusterConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenGaussList) DeepCopyInto(out *OpenGaussList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]OpenGauss, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenGaussList.
func (in *OpenGaussList) DeepCopy() *OpenGaussList {
	if in == nil {
		return nil
	}
	out := new(OpenGaussList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OpenGaussList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenGaussSpec) DeepCopyInto(out *OpenGaussSpec) {
	*out = *in
	if in.OpenGauss != nil {
		in, out := &in.OpenGauss, &out.OpenGauss
		*out = new(OpenGaussClusterConfiguration)
		(*in).DeepCopyInto(*out)
	}
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(corev1.ResourceRequirements)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenGaussSpec.
func (in *OpenGaussSpec) DeepCopy() *OpenGaussSpec {
	if in == nil {
		return nil
	}
	out := new(OpenGaussSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenGaussStatefulSet) DeepCopyInto(out *OpenGaussStatefulSet) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(corev1.ResourceRequirements)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenGaussStatefulSet.
func (in *OpenGaussStatefulSet) DeepCopy() *OpenGaussStatefulSet {
	if in == nil {
		return nil
	}
	out := new(OpenGaussStatefulSet)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenGaussStatus) DeepCopyInto(out *OpenGaussStatus) {
	*out = *in
	if in.MasterIPs != nil {
		in, out := &in.MasterIPs, &out.MasterIPs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ReplicasSmallIPs != nil {
		in, out := &in.ReplicasSmallIPs, &out.ReplicasSmallIPs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ReplicasMidIPs != nil {
		in, out := &in.ReplicasMidIPs, &out.ReplicasMidIPs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ReplicasLargeIPs != nil {
		in, out := &in.ReplicasLargeIPs, &out.ReplicasLargeIPs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenGaussStatus.
func (in *OpenGaussStatus) DeepCopy() *OpenGaussStatus {
	if in == nil {
		return nil
	}
	out := new(OpenGaussStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OriginOpenGaussCluster) DeepCopyInto(out *OriginOpenGaussCluster) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OriginOpenGaussCluster.
func (in *OriginOpenGaussCluster) DeepCopy() *OriginOpenGaussCluster {
	if in == nil {
		return nil
	}
	out := new(OriginOpenGaussCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ShardingsphereStatefulSet) DeepCopyInto(out *ShardingsphereStatefulSet) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(corev1.ResourceRequirements)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ShardingsphereStatefulSet.
func (in *ShardingsphereStatefulSet) DeepCopy() *ShardingsphereStatefulSet {
	if in == nil {
		return nil
	}
	out := new(ShardingsphereStatefulSet)
	in.DeepCopyInto(out)
	return out
}
