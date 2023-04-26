// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package factory

import (
	ccminstall "github.com/verrazzano/verrazzano-modules/ccm-operator/lifecycle-actions/handlers/install"
	compspi "github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/action_spi"
	"github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/handlers/update"
	"github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/handlers/upgrade"

	"github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/handlers/uninstall"
)

// NewLifeCycleComponent creates a new lifecycle component
func NewLifeCycleComponent() compspi.LifecycleComponent {

	// This is an example of how to override just the install lifecycle handler
	return compspi.LifecycleComponent{
		InstallAction:   ccminstall.NewComponent(),
		UninstallAction: uninstall.NewComponent(),
		UpdateAction:    update.NewComponent(),
		UpgradeAction:   upgrade.NewComponent(),
	}
}