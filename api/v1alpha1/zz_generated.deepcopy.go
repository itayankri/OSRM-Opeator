//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2022.

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IngressSpec) DeepCopyInto(out *IngressSpec) {
	*out = *in
	if in.IngressClassName != nil {
		in, out := &in.IngressClassName, &out.IngressClassName
		*out = new(string)
		**out = **in
	}
	if in.ExposingServices != nil {
		in, out := &in.ExposingServices, &out.ExposingServices
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IngressSpec.
func (in *IngressSpec) DeepCopy() *IngressSpec {
	if in == nil {
		return nil
	}
	out := new(IngressSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OSRMCluster) DeepCopyInto(out *OSRMCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OSRMCluster.
func (in *OSRMCluster) DeepCopy() *OSRMCluster {
	if in == nil {
		return nil
	}
	out := new(OSRMCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OSRMCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OSRMClusterList) DeepCopyInto(out *OSRMClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]OSRMCluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OSRMClusterList.
func (in *OSRMClusterList) DeepCopy() *OSRMClusterList {
	if in == nil {
		return nil
	}
	out := new(OSRMClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OSRMClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OSRMClusterSpec) DeepCopyInto(out *OSRMClusterSpec) {
	*out = *in
	in.Profiles.DeepCopyInto(&out.Profiles)
	if in.Ingress != nil {
		in, out := &in.Ingress, &out.Ingress
		*out = new(IngressSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Image != nil {
		in, out := &in.Image, &out.Image
		*out = new(string)
		**out = **in
	}
	in.Persistence.DeepCopyInto(&out.Persistence)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OSRMClusterSpec.
func (in *OSRMClusterSpec) DeepCopy() *OSRMClusterSpec {
	if in == nil {
		return nil
	}
	out := new(OSRMClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OSRMClusterStatus) DeepCopyInto(out *OSRMClusterStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OSRMClusterStatus.
func (in *OSRMClusterStatus) DeepCopy() *OSRMClusterStatus {
	if in == nil {
		return nil
	}
	out := new(OSRMClusterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PersistenceSpec) DeepCopyInto(out *PersistenceSpec) {
	*out = *in
	if in.Storage != nil {
		in, out := &in.Storage, &out.Storage
		x := (*in).DeepCopy()
		*out = &x
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PersistenceSpec.
func (in *PersistenceSpec) DeepCopy() *PersistenceSpec {
	if in == nil {
		return nil
	}
	out := new(PersistenceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProfileSpec) DeepCopyInto(out *ProfileSpec) {
	*out = *in
	if in.MinReplicas != nil {
		in, out := &in.MinReplicas, &out.MinReplicas
		*out = new(int32)
		**out = **in
	}
	if in.MaxReplicas != nil {
		in, out := &in.MaxReplicas, &out.MaxReplicas
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProfileSpec.
func (in *ProfileSpec) DeepCopy() *ProfileSpec {
	if in == nil {
		return nil
	}
	out := new(ProfileSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProfilesSpec) DeepCopyInto(out *ProfilesSpec) {
	*out = *in
	if in.Driving != nil {
		in, out := &in.Driving, &out.Driving
		*out = new(ProfileSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Cycling != nil {
		in, out := &in.Cycling, &out.Cycling
		*out = new(ProfileSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Foot != nil {
		in, out := &in.Foot, &out.Foot
		*out = new(ProfileSpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProfilesSpec.
func (in *ProfilesSpec) DeepCopy() *ProfilesSpec {
	if in == nil {
		return nil
	}
	out := new(ProfilesSpec)
	in.DeepCopyInto(out)
	return out
}
