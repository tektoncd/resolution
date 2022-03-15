// +build !ignore_autogenerated

/*
Copyright 2022 The Tekton Authors

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

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResolutionRequest) DeepCopyInto(out *ResolutionRequest) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResolutionRequest.
func (in *ResolutionRequest) DeepCopy() *ResolutionRequest {
	if in == nil {
		return nil
	}
	out := new(ResolutionRequest)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ResolutionRequest) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResolutionRequestList) DeepCopyInto(out *ResolutionRequestList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ResolutionRequest, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResolutionRequestList.
func (in *ResolutionRequestList) DeepCopy() *ResolutionRequestList {
	if in == nil {
		return nil
	}
	out := new(ResolutionRequestList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ResolutionRequestList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResolutionRequestSpec) DeepCopyInto(out *ResolutionRequestSpec) {
	*out = *in
	if in.Parameters != nil {
		in, out := &in.Parameters, &out.Parameters
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResolutionRequestSpec.
func (in *ResolutionRequestSpec) DeepCopy() *ResolutionRequestSpec {
	if in == nil {
		return nil
	}
	out := new(ResolutionRequestSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResolutionRequestStatus) DeepCopyInto(out *ResolutionRequestStatus) {
	*out = *in
	in.Status.DeepCopyInto(&out.Status)
	out.ResolutionRequestStatusFields = in.ResolutionRequestStatusFields
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResolutionRequestStatus.
func (in *ResolutionRequestStatus) DeepCopy() *ResolutionRequestStatus {
	if in == nil {
		return nil
	}
	out := new(ResolutionRequestStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResolutionRequestStatusFields) DeepCopyInto(out *ResolutionRequestStatusFields) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResolutionRequestStatusFields.
func (in *ResolutionRequestStatusFields) DeepCopy() *ResolutionRequestStatusFields {
	if in == nil {
		return nil
	}
	out := new(ResolutionRequestStatusFields)
	in.DeepCopyInto(out)
	return out
}
