// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package delete

import (
	actionspi "github.com/verrazzano/verrazzano-modules/common/actionspi"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/handlers/common"
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
	return "delete"
}

// Init initializes the handler
func (h *Handler) Init(ctx actionspi.HandlerContext, config actionspi.HandlerConfig) (ctrl.Result, error) {
	return h.BaseHandler.Init(ctx, config, string(moduleapi.ModuleDeleteAction))
}

// IsActionNeeded returns true if delete is needed
func (h Handler) IsActionNeeded(ctx actionspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// PreActionUpdateStatus does the lifecycle pre-Action status update
func (h Handler) PreActionUpdateStatus(ctx actionspi.HandlerContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateStatus(ctx, moduleapi.CondPreUninstall, moduleapi.ModuleStateReconciling)
}

// PreAction does delete pre-action
func (h Handler) PreAction(ctx actionspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// IsPreActionDone returns true if pre-action done
func (h Handler) IsPreActionDone(ctx actionspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// ActionUpdateStatus does the lifecycle Action status update
func (h Handler) ActionUpdateStatus(ctx actionspi.HandlerContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateStatus(ctx, moduleapi.CondUninstallStarted, moduleapi.ModuleStateReconciling)
}

// DoAction uninstalls the component using Helm
func (h Handler) DoAction(ctx actionspi.HandlerContext) (ctrl.Result, error) {
	return h.BaseHandler.DoAction(ctx)
}

// IsActionDone Indicates whether a component is removed
func (h Handler) IsActionDone(ctx actionspi.HandlerContext) (bool, ctrl.Result, error) {
	return h.BaseHandler.IsActionDone(ctx)
}

// PostActionUpdateStatus does delete post-action
func (h Handler) PostActionUpdateStatus(ctx actionspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// PostAction does delete post-action
func (h Handler) PostAction(ctx actionspi.HandlerContext) (ctrl.Result, error) {
	return h.BaseHandler.PostAction(ctx)
}

// IsPostActionDone returns true if post-action done
func (h Handler) IsPostActionDone(ctx actionspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// CompletedActionUpdateStatus does the lifecycle completed Action status update
func (h Handler) CompletedActionUpdateStatus(ctx actionspi.HandlerContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateDoneStatus(ctx, moduleapi.CondUninstallComplete, moduleapi.ModuleStateNone, h.BaseHandler.ModuleCR.Spec.Version)
}
