// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package delete

import (
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/handlers/common"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/handlerspi"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Handler struct {
	common.BaseHandler
}

var (
	_ handlerspi.StateMachineHandler = &Handler{}
)

func NewHandler() handlerspi.StateMachineHandler {
	return &Handler{}
}

// GetActionName returns the action name
func (h Handler) GetWorkName() string {
	return string(moduleapi.DeleteAction)
}

// Init initializes the handler
func (h *Handler) Init(ctx handlerspi.HandlerContext, config handlerspi.StateMachineHandlerConfig) (ctrl.Result, error) {
	return h.BaseHandler.Init(ctx, config, moduleapi.DeleteAction)
}

// PreActionUpdateStatus does the lifecycle pre-Action status update
func (h Handler) PreWorkUpdateStatus(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateStatus(ctx, moduleapi.CondPreUninstall, moduleapi.ModuleStateReconciling)
}

// ActionUpdateStatus does the lifecycle Action status update
func (h Handler) DoWorkUpdateStatus(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateStatus(ctx, moduleapi.CondUninstallStarted, moduleapi.ModuleStateReconciling)
}

// CompletedActionUpdateStatus does the lifecycle completed Action status update
func (h Handler) WorkCompletedUpdateStatus(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateDoneStatus(ctx, moduleapi.CondUninstallComplete, moduleapi.ModuleStateReady, h.BaseHandler.ModuleCR.Spec.Version)
}
