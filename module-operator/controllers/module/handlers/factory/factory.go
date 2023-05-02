// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package factory

import (
	actionspi "github.com/verrazzano/verrazzano-modules/common/actionspi"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/handlers/install"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/handlers/uninstall"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/handlers/update"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/handlers/upgrade"
)

func NewLifecycleActionHandler() actionspi.ActionHandlers {
	return actionspi.ActionHandlers{
		InstallAction:   install.NewHandler(),
		UninstallAction: uninstall.NewHandler(),
		UpdateAction:    update.NewHandler(),
		UpgradeAction:   upgrade.NewHandler(),
	}
}
