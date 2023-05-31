// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package handlerspi

// ModuleHandlerInfo contains the Module handler interfaces.
type ModuleHandlerInfo struct {
	DeleteActionHandler  StateMachineHandler
	InstallActionHandler StateMachineHandler
	UpdateActionHandler  StateMachineHandler
	UpgradeActionHandler StateMachineHandler
}
<<<<<<< HEAD

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
=======
>>>>>>> 31da8ce42cfb64e8d44b0b74f4276a689cdcedce
