// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package install

import (
	compspi "github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/action_spi"
	"github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/handlers/install"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Component struct {
	install.Component
}

func NewComponent() compspi.LifecycleActionHandler {
	return &Component{}
}

// PreAction does installation pre-install
func (h Component) PreAction(context spi.ComponentContext) (ctrl.Result, error) {

	// Do some pre-install work
	// TODO - do your calico specific stuff here

	// Do the common pre-install action
	return h.Component.PreAction(context)
}

// IsPreActionDone returns true if pre-install done
func (h Component) IsPreActionDone(context spi.ComponentContext) (bool, ctrl.Result, error) {

	// Check if the calico pre-install is done
	// TODO - do your calico specific stuff here

	// Do the common method to check if pre-install is done
	return h.Component.IsPreActionDone(context)
}
