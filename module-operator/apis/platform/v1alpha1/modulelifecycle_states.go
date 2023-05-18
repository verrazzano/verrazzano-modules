// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package v1alpha1

// ModuleLifecycleState describes the current reconciling stage of a ModuleLifecycle instance
type ModuleLifecycleState string

const (
	StatePreinstall   ModuleLifecycleState = "PreInstalling"
	StateInstalling   ModuleLifecycleState = "Installing"
	StateUninstalling ModuleLifecycleState = "Uninstalling"
	StatePreUpgrade   ModuleLifecycleState = "PreUpgrading"
	StateUpgrading    ModuleLifecycleState = "Upgrading"
	StateFailed       ModuleLifecycleState = "Failed"
	StateNotNeeded    ModuleLifecycleState = "NotNeeded"
	StateReady        ModuleLifecycleState = "Ready"
	StateCompleted    ModuleLifecycleState = "Completed"
)

type LifecycleCondition string

const (
	ConditionArrayLimit = 5

	CondAlreadyInstalled    LifecycleCondition = "AlreadyInstalled"
	CondAlreadyUninstalled  LifecycleCondition = "AlreadyUninstalled"
	CondAlreadyUpgraded     LifecycleCondition = "AlreadyUpgraded"
	CondPreInstall          LifecycleCondition = "PreInstall"
	CondInstallStarted      LifecycleCondition = "InstallStarted"
	CondInstallComplete     LifecycleCondition = "InstallComplete"
	CondPreUninstall        LifecycleCondition = "PreUninstall"
	CondUninstallStarted    LifecycleCondition = "UninstallStarted"
	CondUninstallComplete   LifecycleCondition = "UninstallComplete"
	CondPreUpgrade          LifecycleCondition = "PreUpgrade"
	CondUpgradeStarted      LifecycleCondition = "UpgradeStarted"
	CondUpgradeComplete     LifecycleCondition = "UpgradeComplete"
	CondReady               LifecycleCondition = "Ready"
	CondReconciling         LifecycleCondition = "Reconciling"
	CondReconcilingComplete LifecycleCondition = "ReconcileComplete"
	CondFailed              LifecycleCondition = "Failed"
)

func (m *ModuleLifecycle) SetState(state ModuleLifecycleState) {
	m.Status.State = state
}

func LifecycleState(condition LifecycleCondition) ModuleLifecycleState {
	switch condition {
	case CondPreInstall:
		return StatePreinstall
	case CondInstallStarted:
		return StateInstalling
	case CondInstallComplete:
		return StateCompleted
	case CondUninstallStarted:
		return StateUninstalling
	case CondUninstallComplete:
		return StateCompleted
	case CondPreUpgrade:
		return StatePreUpgrade
	case CondUpgradeStarted:
		return StateUpgrading
	case CondUpgradeComplete:
		return StateCompleted
	case CondAlreadyInstalled:
		return StateNotNeeded
	case CondAlreadyUninstalled:
		return StateNotNeeded
	case CondFailed:
		return StateFailed
	default:
		return StateReady
	}
}
