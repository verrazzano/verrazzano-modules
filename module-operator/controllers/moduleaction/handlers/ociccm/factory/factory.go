// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package factory

import (
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/moduleaction/handlers/helm/delete"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/moduleaction/handlers/helm/install"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/moduleaction/handlers/helm/update"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/moduleaction/handlers/helm/upgrade"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/handlerspi"
)

// NewModuleActionHandlerInfo creates a new ModuleActionHandlerInfo
func NewModuleActionHandlerInfo() handlerspi.ModuleActionHandlerInfo {
	return handlerspi.ModuleActionHandlerInfo{
		ModuleActualStateInCluster: nil,
		InstallActionHandler:       install.NewHandler(),
		DeleteActionHandler:        delete.NewHandler(),
		UpdateActionHandler:        update.NewHandler(),
		UpgradeActionHandler:       upgrade.NewHandler(),
	}
}