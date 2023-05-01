// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package factory

import (
	compspi "github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/action_spi"
	"github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/handlers/install"
	"github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/handlers/uninstall"
	"github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/handlers/update"
	"github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/handlers/upgrade"
)

func NewLifecycleActionHandler() compspi.ActionHandlers {
	return compspi.ActionHandlers{
		InstallAction:   install.NewComponent(),
		UninstallAction: uninstall.NewComponent(),
		UpdateAction:    update.NewComponent(),
		UpgradeAction:   upgrade.NewComponent(),
	}
}
