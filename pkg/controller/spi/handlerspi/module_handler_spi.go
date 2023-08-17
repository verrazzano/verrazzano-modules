// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package handlerspi

import "github.com/verrazzano/verrazzano-modules/pkg/controller/result"

// MigrationHandler is used when migrating from non-modules on a cluster to modules
type MigrationHandler interface {
	// UpdateStatusIfAlreadyInstalled updates the status if the Module has already been installed
	// without a Module CR by some external actor (such as Verrazzano).
	UpdateStatusIfAlreadyInstalled(context HandlerContext) result.Result
}

// ModuleHandlerInfo contains the Module handler interfaces.
type ModuleHandlerInfo struct {
	DeleteActionHandler  StateMachineHandler
	InstallActionHandler StateMachineHandler
	UpdateActionHandler  StateMachineHandler
	UpgradeActionHandler StateMachineHandler

	// MigrationHandler is used when migrating from non-modules on a cluster to modules. This is optional.
	MigrationHandler
}
