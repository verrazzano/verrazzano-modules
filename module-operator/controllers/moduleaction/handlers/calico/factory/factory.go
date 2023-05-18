// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package factory

import (
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/moduleaction/handlers/calico/install"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/moduleaction/handlers/common"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/moduleaction/handlers/helm/delete"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/moduleaction/handlers/helm/update"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/moduleaction/handlers/helm/upgrade"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/handlerspi"
)

// NewLModuleLifecycleHandlerInfo creates a new ModuleLifecycleHandlerInfo
func NewLModuleLifecycleHandlerInfo() handlerspi.ModuleLifecycleHandlerInfo {
	return handlerspi.ModuleLifecycleHandlerInfo{
		ModuleActualStateInCluster: common.ModuleState{},
		InstallActionHandler:       install.NewHandler(),
		DeleteActionHandler:        delete.NewHandler(),
		UpdateActionHandler:        update.NewHandler(),
		UpgradeActionHandler:       upgrade.NewHandler(),
	}
}
