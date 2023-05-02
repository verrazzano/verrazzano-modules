// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package update

import (
	actionspi "github.com/verrazzano/verrazzano-modules/common/actionspi"
	moduleplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/handlers/common"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Handler struct {
	BaseHandler common.BaseHandler
}

var (
	_ actionspi.LifecycleActionHandler = &Handler{}
)

func NewHandler() actionspi.LifecycleActionHandler {
	return &Handler{}
}

// GetActionName returns the action name
func (h Handler) GetActionName() string {
	return "update"
}

// Init initializes the handler
func (h *Handler) Init(ctx spi.ComponentContext, config actionspi.HandlerConfig) (ctrl.Result, error) {
	return h.BaseHandler.Init(ctx, config, moduleplatform.UpdateAction)
}

// IsActionNeeded returns true if install is needed
func (h Handler) IsActionNeeded(ctx spi.ComponentContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil

	//installed, err := vzhelm.IsReleaseInstalled(h.ReleaseName, h.chartDir)
	//if err != nil {
	//	ctx.Log().ErrorfThrottled("Error checking if Helm release installed for %s/%s", h.chartDir, h.ReleaseName)
	//	return true, ctrl.Result{}, err
	//}
	//return !installed, ctrl.Result{}, err
}

// PreActionUpdateStatus does the lifecycle pre-Action status update
func (h Handler) PreActionUpdateStatus(ctx spi.ComponentContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateStatus(ctx, moduleplatform.CondPreInstall, moduleplatform.ModuleStateReconciling)
}

// PreAction does installation pre-action
func (h Handler) PreAction(ctx spi.ComponentContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// IsPreActionDone returns true if pre-action done
func (h Handler) IsPreActionDone(ctx spi.ComponentContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// ActionUpdateStatus does the lifecycle Action status update
func (h Handler) ActionUpdateStatus(ctx spi.ComponentContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateStatus(ctx, moduleplatform.CondInstallStarted, moduleplatform.ModuleStateReconciling)
}

// DoAction installs the component using Helm
func (h Handler) DoAction(ctx spi.ComponentContext) (ctrl.Result, error) {
	return h.BaseHandler.DoAction(ctx)
}

// IsActionDone Indicates whether a component is installed and ready
func (h Handler) IsActionDone(ctx spi.ComponentContext) (bool, ctrl.Result, error) {
	return h.BaseHandler.IsActionDone(ctx)
}

// PostActionUpdateStatue does installation post-action status update
func (h Handler) PostActionUpdateStatus(ctx spi.ComponentContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// PostAction does installation post-action
func (h Handler) PostAction(ctx spi.ComponentContext) (ctrl.Result, error) {
	return h.BaseHandler.PostAction(ctx)
}

// IsPostActionDone returns true if post-action done
func (h Handler) IsPostActionDone(ctx spi.ComponentContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// CompletedActionUpdateStatus does the lifecycle action complete status update
func (h Handler) CompletedActionUpdateStatus(ctx spi.ComponentContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateStatus(ctx, moduleplatform.CondInstallComplete, moduleplatform.ModuleStateReady)
}
