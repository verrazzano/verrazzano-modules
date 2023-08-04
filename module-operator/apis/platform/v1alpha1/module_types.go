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
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.lastSuccessfulVersion",description="The current version of the Verrazzano platform."
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[:].status",description="True if the Module is ready"
// +kubebuilder:printcolumn:name="Transition Time",type="string",priority=2,JSONPath=".status.conditions[:].lastTransitionTime",description="Last transition time"
// +kubebuilder:printcolumn:name="Message",type="string",priority=2,JSONPath=".status.conditions[:].message",description="Lifecycle message"
// +genclient

// Module specifies a Verrazzano Module instance.
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

// ModuleSpec defines the specification for a Verrazzano Module instance.
type ModuleSpec struct {
	// Module name is the well-known Module name.
	ModuleName string `json:"moduleName,omitempty"`

	// Version is the desired version of the Module.
	// +optional
	Version string `json:"version,omitempty"`

	// TargetNamespace is the namespace where the Module will be installed.
	// +optional
	TargetNamespace string `json:"targetNamespace,omitempty"`

	// Values specifies configuration values using inline YAML.
	// Values have precedence over ValuesFrom.
	// +optional
	Values *apiextensionsv1.JSON `json:"values,omitempty"`

	// ValuesFrom specifies the values from a Configmap or Secret.
	// Each entry in the list has precedence over all previous entries in the list.
	// +optional
	ValuesFrom []ValuesFromSource `json:"valuesFrom,omitempty"`
}

// ValuesFromSource specifies value overrides for a Module.
type ValuesFromSource struct {
	// ConfigMapRef is a selector for a ConfigMap containing values data.
	// +optional
	ConfigMapRef *corev1.ConfigMapKeySelector `json:"configMapRef,omitempty"`

	// SecretRef is a selector for a Secret containing values data.
	// +optional
	SecretRef *corev1.SecretKeySelector `json:"secretRef,omitempty"`
}

// ModuleConditionType is the condition type.
type ModuleConditionType string

const (
	// ModuleConditionReady is the Ready condition type
	ModuleConditionReady = "Ready"
)

// ModuleStatus defines the action state of the Module resource.
type ModuleStatus struct {
	// Conditions are the list of conditions for the Module.
	Conditions []ModuleCondition `json:"conditions,omitempty"`

	// LastSuccessfulVersion is the last version of the Module that was successfully reconciled.
	// +optional
	LastSuccessfulVersion string `json:"lastSuccessfulVersion,omitempty"`

	// LastSuccessfulGeneration is the last generation of the Module that was successfully reconciled.
	// +optional
	LastSuccessfulGeneration int64 `json:"lastSuccessfulGeneration,omitempty"`
}

// ModuleCondition describes the current condition of the Module.
type ModuleCondition struct {
	// LastTransitionTime is the last time the condition transitioned from one status to another.
	LastTransitionTime string `json:"lastTransitionTime"`

	// Message is a human-readable message indicating details about the last transition.
	Message string `json:"message,omitempty"`

	// Status of the condition: one of `True`, `False`, or `Unknown`.
	Status corev1.ConditionStatus `json:"status"`

	// Type of condition.
	Type ModuleConditionType `json:"type"`

	// Reason for the condition.  This is a machine-readable one word value.
	Reason ModuleConditionReason `json:"reason"`
}

// ModuleClassType Identifies the class used to manage a set of Module types.
type ModuleClassType string

const (
	// HelmModuleClass defines the class type used by the Helm operator.
	HelmModuleClass ModuleClassType = "helm"

	// CalicoModuleClass defines the class type used by the Calico operator.
	CalicoModuleClass ModuleClassType = "calico"

	// CCMModuleClass defines the class type used by the oci-ccm operator.
	CCMModuleClass ModuleClassType = "oci-ccm"

	// MultusModuleClass defines the class type used by the multus operator.
	MultusModuleClass ModuleClassType = "multus"

	// MetallbModuleClass defines the class type used by the Metallb operator.
	MetallbModuleClass ModuleClassType = "metallb"
)

// ModuleConditionReason is the reason for the condition type.
type ModuleConditionReason string

const (
	ReadyReasonInstallStarted     ModuleConditionReason = "InstallStarted"
	ReadyReasonInstallSucceeded   ModuleConditionReason = "InstallSucceeded"
	ReadyReasonInstallFailed      ModuleConditionReason = "InstallFailed"
	ReadyReasonUninstallStarted   ModuleConditionReason = "UninstallStarted"
	ReadyReasonUninstallSucceeded ModuleConditionReason = "UninstallSucceeded"
	ReadyReasonUninstallFailed    ModuleConditionReason = "UninstallFailed"
	ReadyReasonUpdateStarted      ModuleConditionReason = "UpdateStarted"
	ReadyReasonUpdateSucceeded    ModuleConditionReason = "UpdateSucceeded"
	ReadyReasonUpdateFailed       ModuleConditionReason = "UpdateFailed"
	ReadyReasonUpgradeStarted     ModuleConditionReason = "UpgradeStarted"
	ReadyReasonUpgradeSucceeded   ModuleConditionReason = "UpgradeSucceeded"
	ReadyReasonUpgradeFailed      ModuleConditionReason = "UpgradeFailed"
)

func init() {
	SchemeBuilder.Register(&Module{}, &ModuleList{})
}
