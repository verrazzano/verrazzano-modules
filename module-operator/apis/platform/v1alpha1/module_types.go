// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:path=modules,shortName=module;modules
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.version",description="The current version of the Verrazzano platform."
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.state",description="State of Module reconciliation"
// +genclient

// Module specifies a Verrazzano Module instance
type Module struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModuleSpec   `json:"spec,omitempty"`
	Status ModuleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ModuleList contains a list of Verrazzano Module instance resources.
type ModuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Module `json:"items"`
}

// ModuleSpec defines the attributes for a Verrazzano Module instance
type ModuleSpec struct {
	ModuleName      string      `json:"moduleName,omitempty"`
	Version         string      `json:"version,omitempty"`
	TargetNamespace string      `json:"targetNamespace,omitempty"`
	Overrides       []Overrides `json:"overrides,omitempty"`
}

// Overrides identifies overrides for a component.
type Overrides struct {
	// Selector for ConfigMap containing override data.
	// +optional
	ConfigMapRef *corev1.ConfigMapKeySelector `json:"configMapRef,omitempty"`
	// Selector for Secret containing override data.
	// +optional
	SecretRef *corev1.SecretKeySelector `json:"secretRef,omitempty"`
	// Configure overrides using inline YAML.
	// +optional
	Values *apiextensionsv1.JSON `json:"values,omitempty"`
}

type ModuleStateType string

const (
	ModuleStateReady       = "Ready"
	ModuleStateReconciling = "Reconciling"
)

// ModuleStatus defines the observed state of a Verrazzano Module resource.
type ModuleStatus struct {
	// State is the Module state
	State ModuleStateType `json:"state,omitempty"`
	// The latest available observations of an object's current state.
	Conditions []ModuleCondition `json:"conditions,omitempty"`
	// ObservedGeneration is the actual generation that was reconciled
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// The version of module that is installed.
	Version string `json:"version,omitempty"`
}

// ModuleCondition describes the current state of an installation.
type ModuleCondition struct {
	// Last time the condition transitioned from one status to another.
	LastTransitionTime string `json:"lastTransitionTime,omitempty"`
	// Human-readable message indicating details about the last transition.
	Message string `json:"message,omitempty"`
	// Status of the condition: one of `True`, `False`, or `Unknown`.
	Status corev1.ConditionStatus `json:"status"`
	// Type of condition.
	Type LifecycleCondition `json:"type"`
}

// ModuleClassType Identifies the lifecycle class used to manage a subset of Module types
type ModuleClassType string

const (
	// HelmModuleClass defines the class name used by the Helm operator
	HelmModuleClass ModuleClassType = "helm"

	// CalicoModuleClass defines the class name used by the Calico operator
	CalicoModuleClass ModuleClassType = "calico"
	CCMModuleClass    ModuleClassType = "oci-ccm"
)

type LifecycleCondition string

const (
	ConditionArrayLimit = 5

	CondPreInstall        LifecycleCondition = "PreInstall"
	CondInstallStarted    LifecycleCondition = "InstallStarted"
	CondInstallComplete   LifecycleCondition = "InstallComplete"
	CondPreUninstall      LifecycleCondition = "PreUninstall"
	CondUninstallStarted  LifecycleCondition = "UninstallStarted"
	CondUninstallComplete LifecycleCondition = "UninstallComplete"
	CondPreUpgrade        LifecycleCondition = "PreUpgrade"
	CondUpgradeStarted    LifecycleCondition = "UpgradeStarted"
	CondUpgradeComplete   LifecycleCondition = "UpgradeComplete"
)

func init() {
	SchemeBuilder.Register(&Module{}, &ModuleList{})
}
