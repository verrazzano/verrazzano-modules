// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package upgrade

import (
	actionspi "github.com/verrazzano/verrazzano-modules/common/actionspi"
	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"

	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
)
import (
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/handlers/common"
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
	return "upgrade"
}

// Init initializes the handler
func (h *Handler) Init(ctx spi.ComponentContext, config actionspi.HandlerConfig) (ctrl.Result, error) {
	return h.BaseHandler.Init(ctx, config, moduleapi.UpgradeAction)
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
	return h.BaseHandler.UpdateStatus(ctx, moduleapi.CondPreUpgrade, moduleapi.ModuleStateReconciling)
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
	return h.BaseHandler.UpdateStatus(ctx, moduleapi.CondUpgradeStarted, moduleapi.ModuleStateReconciling)
}

// DoAction installs the component using Helm
func (h Handler) DoAction(ctx spi.ComponentContext) (ctrl.Result, error) {
	return h.BaseHandler.DoAction(ctx)
}

// IsActionDone Indicates whether a component is installed and ready
func (h Handler) IsActionDone(ctx spi.ComponentContext) (bool, ctrl.Result, error) {
	if ctx.IsDryRun() {
		return true, ctrl.Result{}, nil
	}

	mlc, err := h.BaseHandler.GetModuleLifecycle(ctx)
	if err != nil {
		return false, util.NewRequeueWithShortDelay(), nil
	}
	if mlc.Status.State == moduleapi.StateReady || mlc.Status.State == moduleapi.StateCompleted || mlc.Status.State == moduleapi.StateNotNeeded {
		return true, ctrl.Result{}, nil
	}
	ctx.Log().Progressf("Waiting for ModuleLifecycle %s to be completed", h.BaseHandler.MlcName)
	return false, ctrl.Result{}, nil
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

// CompletedActionUpdateStatus does the lifecycle pre-Action status update
func (h Handler) CompletedActionUpdateStatus(ctx spi.ComponentContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateStatusWithVersion(ctx, moduleapi.CondUpgradeComplete, moduleapi.ModuleStateReady, h.BaseHandler.ModuleCR.Spec.Version)
}
