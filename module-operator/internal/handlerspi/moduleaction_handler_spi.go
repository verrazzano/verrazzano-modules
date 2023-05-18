// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package handlerspi

import (
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// ModuleLifecycleHandlerInfo contains the ModuleAction handler interfaces.
type ModuleLifecycleHandlerInfo struct {
	ModuleActualStateInCluster
	DeleteActionHandler  StateMachineHandler
	InstallActionHandler StateMachineHandler
	UpdateActionHandler  StateMachineHandler
	UpgradeActionHandler StateMachineHandler
}

// ModuleActualState is the actual state of the module in the cluster
type ModuleActualState string

const (
	// ModuleStateFailed means the module is failed
	ModuleStateFailed ModuleActualState = "Failed"

	// ModuleStateNotInstalled means the module is not installed
	ModuleStateNotInstalled ModuleActualState = "NotInstalled"

	// ModuleStateReady means the module is installed or upgraded and is in a ready state
	ModuleStateReady ModuleActualState = "Ready"

	// ModuleStateReconciling means the module installation, upgrade, or deletion is in progress
	ModuleStateReconciling ModuleActualState = "Reconciling"

	// ModuleStateUnknown means the module is unknown
	ModuleStateUnknown ModuleActualState = "Unknown"
)

// ModuleActualStateInCluster interface describes the actual state of the module in the cluster
type ModuleActualStateInCluster interface {
	// GetActualModuleState gets the state of the module
	GetActualModuleState(context HandlerContext, cr *moduleapi.ModuleAction) (ModuleActualState, ctrl.Result, error)

	// IsUpgradeNeeded checks if upgrade is needed
	IsUpgradeNeeded(context HandlerContext, cr *moduleapi.ModuleAction) (bool, ctrl.Result, error)
}
