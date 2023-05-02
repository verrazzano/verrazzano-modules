// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package install

import (
	actionspi "github.com/verrazzano/verrazzano-modules/common/actionspi"
	"github.com/verrazzano/verrazzano-modules/common/controllers/modulelifecycle/handlers/install"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Handler struct {
	InstallHandler install.Handler
}

var (
	_ actionspi.LifecycleActionHandler = &Handler{}
)

func NewComponent() actionspi.LifecycleActionHandler {
	return &Handler{}
}

// PreAction does installation pre-install
func (h Handler) PreAction(ctx spi.ComponentContext) (ctrl.Result, error) {

	// Do some pre-install work
	// TODO - do your calico specific stuff here

	// Do the common pre-install action
	return h.InstallHandler.PreAction(ctx)
}

// IsPreActionDone returns true if pre-install done
func (h Handler) IsPreActionDone(ctx spi.ComponentContext) (bool, ctrl.Result, error) {

	// Check if the calico pre-install is done
	// TODO - do your calico specific stuff here

	// Do the common method to check if pre-install is done
	return h.InstallHandler.IsPreActionDone(ctx)
}

func (h Handler) GetActionName() string {
	return h.InstallHandler.GetActionName()
}

func (h Handler) Init(ctx spi.ComponentContext, config actionspi.HandlerConfig) (ctrl.Result, error) {
	return h.InstallHandler.Init(ctx, config)
}

func (h Handler) IsActionNeeded(ctx spi.ComponentContext) (bool, ctrl.Result, error) {
	return h.InstallHandler.IsActionNeeded(ctx)
}

func (h Handler) PreActionUpdateStatus(ctx spi.ComponentContext) (ctrl.Result, error) {
	return h.InstallHandler.PreActionUpdateStatus(ctx)
}

func (h Handler) ActionUpdateStatus(ctx spi.ComponentContext) (ctrl.Result, error) {
	return h.InstallHandler.ActionUpdateStatus(ctx)
}

func (h Handler) DoAction(ctx spi.ComponentContext) (ctrl.Result, error) {
	// TODO - do your calico specific stuff here

	return h.InstallHandler.DoAction(ctx)
}

func (h Handler) IsActionDone(ctx spi.ComponentContext) (bool, ctrl.Result, error) {
	return h.InstallHandler.IsActionDone(ctx)
}

func (h Handler) PostActionUpdateStatus(ctx spi.ComponentContext) (ctrl.Result, error) {
	return h.InstallHandler.PostActionUpdateStatus(ctx)
}

func (h Handler) PostAction(ctx spi.ComponentContext) (ctrl.Result, error) {
	return h.InstallHandler.PostAction(ctx)
}

func (h Handler) IsPostActionDone(ctx spi.ComponentContext) (bool, ctrl.Result, error) {
	return h.InstallHandler.IsPostActionDone(ctx)
}

func (h Handler) CompletedActionUpdateStatus(ctx spi.ComponentContext) (ctrl.Result, error) {
	return h.InstallHandler.CompletedActionUpdateStatus(ctx)
}
