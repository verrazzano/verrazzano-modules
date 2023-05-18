// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package reconcile

import (
	"context"
	actionspi "github.com/verrazzano/verrazzano-modules/common/handlerspi"
	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/handlers/common"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"

	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
)

type Handler struct {
	BaseHandler common.BaseHandler
}

var (
	_ actionspi.StateMachineHandler = &Handler{}
)

func NewHandler() actionspi.StateMachineHandler {
	return &Handler{}
}

// GetActionName returns the action name
func (h Handler) GetActionName() string {
	return string(moduleapi.ReconcileAction)
}

// Init initializes the handler
func (h *Handler) Init(ctx actionspi.HandlerContext, config actionspi.StateMachineHandlerConfig) (ctrl.Result, error) {
	return h.BaseHandler.Init(ctx, config, moduleapi.ReconcileAction)
}

// IsActionNeeded returns true if install is needed
func (h Handler) IsActionNeeded(ctx actionspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// PreActionUpdateStatus does the lifecycle pre-Action status update
func (h Handler) PreActionUpdateStatus(ctx actionspi.HandlerContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateStatus(ctx, moduleapi.CondPreInstall, moduleapi.ModuleStateReconciling)
}

// PreAction does installation pre-action
func (h Handler) PreAction(ctx actionspi.HandlerContext) (ctrl.Result, error) {
	// Update the spec version if it is not set
	if len(h.BaseHandler.ModuleCR.Spec.Version) == 0 {
		// Update spec version to match chart, always requeue to get ModuleCR with version
		h.BaseHandler.ModuleCR.Spec.Version = h.BaseHandler.Config.ChartInfo.Version
		if err := ctx.Client.Update(context.TODO(), h.BaseHandler.ModuleCR); err != nil {
			return util.NewRequeueWithShortDelay(), nil
		}
		// ALways reconcile so that we get a new tracker with the latest ModuleCR
		return util.NewRequeueWithDelay(1, 2, time.Second), nil
	}

	return ctrl.Result{}, nil
}

// IsPreActionDone returns true if pre-action done
func (h Handler) IsPreActionDone(ctx actionspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// ActionUpdateStatus does the lifecycle Action status update
func (h Handler) ActionUpdateStatus(ctx actionspi.HandlerContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateStatus(ctx, moduleapi.CondInstallStarted, moduleapi.ModuleStateReconciling)
}

// DoAction installs the component using Helm
func (h Handler) DoAction(ctx actionspi.HandlerContext) (ctrl.Result, error) {
	return h.BaseHandler.DoAction(ctx)
}

// IsActionDone Indicates whether a component is installed and ready
func (h Handler) IsActionDone(ctx actionspi.HandlerContext) (bool, ctrl.Result, error) {
	return h.BaseHandler.IsActionDone(ctx)
}

// PostActionUpdateStatue does installation post-action status update
func (h Handler) PostActionUpdateStatus(ctx actionspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// PostAction does installation post-action
func (h Handler) PostAction(ctx actionspi.HandlerContext) (ctrl.Result, error) {
	return h.BaseHandler.PostAction(ctx)
}

// IsPostActionDone returns true if post-action done
func (h Handler) IsPostActionDone(ctx actionspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// CompletedActionUpdateStatus does the lifecycle pre-Action status update
func (h Handler) CompletedActionUpdateStatus(ctx actionspi.HandlerContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateDoneStatus(ctx, moduleapi.CondInstallComplete, moduleapi.ModuleStateReady, h.BaseHandler.ModuleCR.Spec.Version)
}
