// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package factory

import (
	actionspi "github.com/verrazzano/verrazzano-modules/common/actionspi"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/modulelifecycle/handlers/helm/install"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/modulelifecycle/handlers/helm/uninstall"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/modulelifecycle/handlers/helm/update"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/modulelifecycle/handlers/helm/upgrade"
)

func NewLifecycleActionHandler() actionspi.ActionHandlers {
	return actionspi.ActionHandlers{
		InstallActionHandler:   install.NewHandler(),
		UninstallActionHandler: uninstall.NewHandler(),
		UpdateActionHandler:    update.NewHandler(),
		UpgradeActionHandler:   upgrade.NewHandler(),
	}
}