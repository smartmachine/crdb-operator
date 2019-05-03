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

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CockroachDB) DeepCopyInto(out *CockroachDB) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CockroachDB.
func (in *CockroachDB) DeepCopy() *CockroachDB {
	if in == nil {
		return nil
	}
	out := new(CockroachDB)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CockroachDB) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CockroachDBClientSpec) DeepCopyInto(out *CockroachDBClientSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CockroachDBClientSpec.
func (in *CockroachDBClientSpec) DeepCopy() *CockroachDBClientSpec {
	if in == nil {
		return nil
	}
	out := new(CockroachDBClientSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CockroachDBClusterSpec) DeepCopyInto(out *CockroachDBClusterSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CockroachDBClusterSpec.
func (in *CockroachDBClusterSpec) DeepCopy() *CockroachDBClusterSpec {
	if in == nil {
		return nil
	}
	out := new(CockroachDBClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CockroachDBDashboardSpec) DeepCopyInto(out *CockroachDBDashboardSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CockroachDBDashboardSpec.
func (in *CockroachDBDashboardSpec) DeepCopy() *CockroachDBDashboardSpec {
	if in == nil {
		return nil
	}
	out := new(CockroachDBDashboardSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CockroachDBList) DeepCopyInto(out *CockroachDBList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CockroachDB, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CockroachDBList.
func (in *CockroachDBList) DeepCopy() *CockroachDBList {
	if in == nil {
		return nil
	}
	out := new(CockroachDBList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CockroachDBList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CockroachDBNode) DeepCopyInto(out *CockroachDBNode) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CockroachDBNode.
func (in *CockroachDBNode) DeepCopy() *CockroachDBNode {
	if in == nil {
		return nil
	}
	out := new(CockroachDBNode)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CockroachDBSpec) DeepCopyInto(out *CockroachDBSpec) {
	*out = *in
	out.Cluster = in.Cluster
	out.Client = in.Client
	out.Dashboard = in.Dashboard
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CockroachDBSpec.
func (in *CockroachDBSpec) DeepCopy() *CockroachDBSpec {
	if in == nil {
		return nil
	}
	out := new(CockroachDBSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CockroachDBStatus) DeepCopyInto(out *CockroachDBStatus) {
	*out = *in
	if in.Nodes != nil {
		in, out := &in.Nodes, &out.Nodes
		*out = make([]CockroachDBNode, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CockroachDBStatus.
func (in *CockroachDBStatus) DeepCopy() *CockroachDBStatus {
	if in == nil {
		return nil
	}
	out := new(CockroachDBStatus)
	in.DeepCopyInto(out)
	return out
}