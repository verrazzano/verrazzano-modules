// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.state",description="State of Module reconciliation"
//+kubebuilder:storageversion
//+kubebuilder:resource:path=modulelifecycles,shortName=mlc;mlcs
//+genclient

// ModuleLifecycle defines the schema for a module lifecycle operation
type ModuleLifecycle struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModuleLifecycleSpec   `json:"spec,omitempty"`
	Status ModuleLifecycleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ModuleLifecycleList contains a list of Module
type ModuleLifecycleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ModuleLifecycle `json:"items"`
}

type ModuleLifecycleSpec struct {
	// LifecycleClassName defines the lifecycle class name required to process the ModuleLifecycle instance
	LifecycleClassName LifecycleClassType `json:"lifecycleClassName,omitempty"`
	// Action Defines lifecycle action to perform
	Action ModuleLifecycleActionType `json:"action"`
	// Installer Defines the installer information required to perform the lifecycle operation
	Installer ModuleInstaller `json:"installer"`
	// The Module version
	Version string `json:"version,omitempty"`
}

// ModuleInstaller Defines the installer information for a module; only one of the fields can be set
type ModuleInstaller struct {
	HelmRelease *HelmRelease `json:"helmRelease,omitempty"`
}

type HelmChart struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Path    string `json:"path,omitempty"`
}

type HelmRelease struct {
	Name       string              `json:"name"`
	Namespace  string              `json:"namespace,omitempty"`
	ChartInfo  HelmChart           `json:"chart,omitempty"`
	Repository HelmChartRepository `json:"repo,omitempty"`
	Overrides  []Overrides         `json:"overrides,omitempty"`
}

// ModuleLifecycleStatus defines the observed state of Module
type ModuleLifecycleStatus struct {
	// Information about the current state of a component
	State              ModuleLifecycleState       `json:"state,omitempty"`
	Conditions         []ModuleLifecycleCondition `json:"conditions,omitempty"`
	ObservedGeneration int64                      `json:"observedGeneration,omitempty"`
	ReconciledAt       string                     `json:"reconciledAt,omitempty"`
	// The Module version
	Version string `json:"version,omitempty"`
}

// ModuleLifecycleCondition describes current state of an installation.
type ModuleLifecycleCondition struct {
	// Type of condition.
	Type LifecycleCondition `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime string `json:"lastTransitionTime,omitempty"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

// LifecycleClassType Identifies the lifecycle class used to manage a subset of Module types
type LifecycleClassType string

const (
	// HelmLifecycleClass defines the class name used by the Helm operator
	HelmLifecycleClass LifecycleClassType = "helm"

	// CalicoLifecycleClass defines the class name used by the Calico operator
	CalicoLifecycleClass LifecycleClassType = "calico"
	CCMLifecycleClass    LifecycleClassType = "ccm"
)

// ModuleLifecycleActionType defines the type of action to be performed in a ModuleLifecycle instance
type ModuleLifecycleActionType string

const (
	// ReconcileAction indicates the ModuleLifecycle CR should reconcile the module
	ReconcileAction ModuleLifecycleActionType = "reconcile"

	// DeleteAction indicates the ModuleLifecycle CR should delete the module
	DeleteAction ModuleLifecycleActionType = "delete"
)

func init() {
	SchemeBuilder.Register(&ModuleLifecycle{}, &ModuleLifecycleList{})
}

func (m *ModuleLifecycle) ChartNamespace() string {
	if m.Spec.Installer.HelmRelease == nil {
		return m.Namespace
	}
	if m.Spec.Installer.HelmRelease != nil && len(m.Spec.Installer.HelmRelease.Namespace) > 0 {
		return m.Spec.Installer.HelmRelease.Namespace
	}
	return "default"
}

func (m *ModuleLifecycle) GetReleaseName() string {
	helmRelease := m.Spec.Installer.HelmRelease
	if helmRelease != nil && len(helmRelease.Name) > 0 {
		return helmRelease.Name
	}
	return m.Name
}

func (m *ModuleLifecycle) IsBeingDeleted() bool {
	return m != nil && m.GetDeletionTimestamp() != nil
}
