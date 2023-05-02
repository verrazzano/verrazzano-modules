// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package factory

import (
	calicoinstall "github.com/verrazzano/verrazzano-modules/calico-operator/lifecycle-actions/handlers/install"
	actionspi "github.com/verrazzano/verrazzano-modules/common/actionspi"
	"github.com/verrazzano/verrazzano-modules/common/controllers/modulelifecycle/handlers/uninstall"
	"github.com/verrazzano/verrazzano-modules/common/controllers/modulelifecycle/handlers/update"
	"github.com/verrazzano/verrazzano-modules/common/controllers/modulelifecycle/handlers/upgrade"
)

func NewLifecycleActionHandler() actionspi.ActionHandlers {
	return actionspi.ActionHandlers{
		InstallAction:   calicoinstall.NewComponent(),
		UninstallAction: uninstall.NewComponent(),
		UpdateAction:    update.NewComponent(),
		UpgradeAction:   upgrade.NewComponent(),
	}
}
