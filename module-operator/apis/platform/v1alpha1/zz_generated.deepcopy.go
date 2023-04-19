//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ChartVersion) DeepCopyInto(out *ChartVersion) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ChartVersion.
func (in *ChartVersion) DeepCopy() *ChartVersion {
	if in == nil {
		return nil
	}
	out := new(ChartVersion)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HelmChart) DeepCopyInto(out *HelmChart) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HelmChart.
func (in *HelmChart) DeepCopy() *HelmChart {
	if in == nil {
		return nil
	}
	out := new(HelmChart)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HelmChartRepository) DeepCopyInto(out *HelmChartRepository) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HelmChartRepository.
func (in *HelmChartRepository) DeepCopy() *HelmChartRepository {
	if in == nil {
		return nil
	}
	out := new(HelmChartRepository)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HelmRelease) DeepCopyInto(out *HelmRelease) {
	*out = *in
	out.ChartInfo = in.ChartInfo
	out.Repository = in.Repository
	if in.Overrides != nil {
		in, out := &in.Overrides, &out.Overrides
		*out = make([]Overrides, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HelmRelease.
func (in *HelmRelease) DeepCopy() *HelmRelease {
	if in == nil {
		return nil
	}
	out := new(HelmRelease)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Module) DeepCopyInto(out *Module) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Module.
func (in *Module) DeepCopy() *Module {
	if in == nil {
		return nil
	}
	out := new(Module)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Module) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModuleCondition) DeepCopyInto(out *ModuleCondition) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModuleCondition.
func (in *ModuleCondition) DeepCopy() *ModuleCondition {
	if in == nil {
		return nil
	}
	out := new(ModuleCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModuleInstaller) DeepCopyInto(out *ModuleInstaller) {
	*out = *in
	if in.HelmRelease != nil {
		in, out := &in.HelmRelease, &out.HelmRelease
		*out = new(HelmRelease)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModuleInstaller.
func (in *ModuleInstaller) DeepCopy() *ModuleInstaller {
	if in == nil {
		return nil
	}
	out := new(ModuleInstaller)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModuleLifecycle) DeepCopyInto(out *ModuleLifecycle) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModuleLifecycle.
func (in *ModuleLifecycle) DeepCopy() *ModuleLifecycle {
	if in == nil {
		return nil
	}
	out := new(ModuleLifecycle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ModuleLifecycle) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModuleLifecycleCondition) DeepCopyInto(out *ModuleLifecycleCondition) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModuleLifecycleCondition.
func (in *ModuleLifecycleCondition) DeepCopy() *ModuleLifecycleCondition {
	if in == nil {
		return nil
	}
	out := new(ModuleLifecycleCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModuleLifecycleList) DeepCopyInto(out *ModuleLifecycleList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ModuleLifecycle, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModuleLifecycleList.
func (in *ModuleLifecycleList) DeepCopy() *ModuleLifecycleList {
	if in == nil {
		return nil
	}
	out := new(ModuleLifecycleList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ModuleLifecycleList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModuleLifecycleSpec) DeepCopyInto(out *ModuleLifecycleSpec) {
	*out = *in
	in.Installer.DeepCopyInto(&out.Installer)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModuleLifecycleSpec.
func (in *ModuleLifecycleSpec) DeepCopy() *ModuleLifecycleSpec {
	if in == nil {
		return nil
	}
	out := new(ModuleLifecycleSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModuleLifecycleStatus) DeepCopyInto(out *ModuleLifecycleStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]ModuleLifecycleCondition, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModuleLifecycleStatus.
func (in *ModuleLifecycleStatus) DeepCopy() *ModuleLifecycleStatus {
	if in == nil {
		return nil
	}
	out := new(ModuleLifecycleStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModuleList) DeepCopyInto(out *ModuleList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Module, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModuleList.
func (in *ModuleList) DeepCopy() *ModuleList {
	if in == nil {
		return nil
	}
	out := new(ModuleList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ModuleList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModuleSource) DeepCopyInto(out *ModuleSource) {
	*out = *in
	if in.ChartRepo != nil {
		in, out := &in.ChartRepo, &out.ChartRepo
		*out = new(HelmChartRepository)
		**out = **in
	}
	if in.SourceRef != nil {
		in, out := &in.SourceRef, &out.SourceRef
		*out = new(ModuleSourceRef)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModuleSource.
func (in *ModuleSource) DeepCopy() *ModuleSource {
	if in == nil {
		return nil
	}
	out := new(ModuleSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModuleSourceRef) DeepCopyInto(out *ModuleSourceRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModuleSourceRef.
func (in *ModuleSourceRef) DeepCopy() *ModuleSourceRef {
	if in == nil {
		return nil
	}
	out := new(ModuleSourceRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModuleSpec) DeepCopyInto(out *ModuleSpec) {
	*out = *in
	if in.Source != nil {
		in, out := &in.Source, &out.Source
		*out = new(ModuleSource)
		(*in).DeepCopyInto(*out)
	}
	if in.Overrides != nil {
		in, out := &in.Overrides, &out.Overrides
		*out = make([]Overrides, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModuleSpec.
func (in *ModuleSpec) DeepCopy() *ModuleSpec {
	if in == nil {
		return nil
	}
	out := new(ModuleSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModuleStatus) DeepCopyInto(out *ModuleStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]ModuleCondition, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModuleStatus.
func (in *ModuleStatus) DeepCopy() *ModuleStatus {
	if in == nil {
		return nil
	}
	out := new(ModuleStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Overrides) DeepCopyInto(out *Overrides) {
	*out = *in
	if in.ConfigMapRef != nil {
		in, out := &in.ConfigMapRef, &out.ConfigMapRef
		*out = new(v1.ConfigMapKeySelector)
		(*in).DeepCopyInto(*out)
	}
	if in.SecretRef != nil {
		in, out := &in.SecretRef, &out.SecretRef
		*out = new(v1.SecretKeySelector)
		(*in).DeepCopyInto(*out)
	}
	if in.Values != nil {
		in, out := &in.Values, &out.Values
		*out = new(apiextensionsv1.JSON)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Overrides.
func (in *Overrides) DeepCopy() *Overrides {
	if in == nil {
		return nil
	}
	out := new(Overrides)
	in.DeepCopyInto(out)
	return out
}
