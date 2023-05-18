// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package handlerspi

import ctrl "sigs.k8s.io/controller-runtime"

// ModuleLifecycleHandlerInfo contains the ModuleLifecycle handler interfaces.
type ModuleLifecycleHandlerInfo struct {
	ModuleActualState
	InstallActionHandler   StateMachineHandler
	UninstallActionHandler StateMachineHandler
	UpdateActionHandler    StateMachineHandler
	UpgradeActionHandler   StateMachineHandler
}

// ModuleActualState interface describes the actual state of the module in the cluster
type ModuleActualState interface {
	// IsInstalled returns true if the module is installed
	IsInstalled(context HandlerContext) (bool, ctrl.Result, error)

	// IsInstallInProgress returns true if the module is being installed
	IsInstallInProgress(context HandlerContext) (bool, ctrl.Result, error)
}
