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
