// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package factory

import (
	"github.com/verrazzano/verrazzano-modules/common/helm_component/lifecycle/install"
	compspi "github.com/verrazzano/verrazzano-modules/common/helm_component/spi"
)

func NewLifeCycleComponent() compspi.LifecycleComponent {
	return compspi.LifecycleComponent{
		InstallAction:   install.NewComponent(),
		UninstallAction: install.NewComponent(),
		UpdateAction:    install.NewComponent(),
		UpgradeAction:   install.NewComponent(),
	}
}
