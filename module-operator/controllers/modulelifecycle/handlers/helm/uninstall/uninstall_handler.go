// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package uninstall

import (
	actionspi "github.com/verrazzano/verrazzano-modules/common/handlerspi"
	"github.com/verrazzano/verrazzano-modules/common/pkg/helm"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/modulelifecycle/handlers/common"
	ctrl "sigs.k8s.io/controller-runtime"
)

type HelmHandler struct {
	common.BaseHandler
}

var (
	_ actionspi.StateMachineHandler = &HelmHandler{}
)

func NewHandler() actionspi.StateMachineHandler {
	return &HelmHandler{}
}

// Init initializes the handler with Helm chart information
func (h *HelmHandler) Init(ctx actionspi.HandlerContext, config actionspi.StateMachineHandlerConfig) (ctrl.Result, error) {
	return h.BaseHandler.Init(ctx, config)
}

// GetActionName returns the action name
func (h HelmHandler) GetActionName() string {
	return "uninstall"
}

// IsActionNeeded returns true if uninstall is needed
func (h HelmHandler) IsActionNeeded(context actionspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// PreActionUpdateStatus does the lifecycle pre-Action status update
func (h HelmHandler) PreActionUpdateStatus(ctx actionspi.HandlerContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateStatus(ctx, moduleapi.CondPreUninstall, moduleapi.ModuleStateReconciling)
}

// PreAction does installation pre-action
func (h HelmHandler) PreAction(ctx actionspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// IsPreActionDone returns true if pre-action done
func (h HelmHandler) IsPreActionDone(ctx actionspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// ActionUpdateStatus does the lifecycle Action status update
func (h HelmHandler) ActionUpdateStatus(ctx actionspi.HandlerContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateStatus(ctx, moduleapi.CondUninstallStarted, moduleapi.ModuleStateReconciling)
}

// DoAction uninstalls the module using Helm
func (h HelmHandler) DoAction(context actionspi.HandlerContext) (ctrl.Result, error) {
	installed, err := helm.IsReleaseInstalled(h.HelmRelease.Name, h.HelmRelease.Namespace)
	if err != nil {
		context.Log.ErrorfThrottled("Error checking if Helm release installed for %s/%s", h.HelmRelease.Namespace, h.HelmRelease.Name)
		return ctrl.Result{}, err
	}
	if !installed {
		return ctrl.Result{}, err
	}

	err = helm.Uninstall(context.Log, h.HelmRelease.Name, h.HelmRelease.Namespace, context.DryRun)
	return ctrl.Result{}, err
}

// IsActionDone Indicates whether a module is uninstalled
func (h HelmHandler) IsActionDone(context actionspi.HandlerContext) (bool, ctrl.Result, error) {
	if context.DryRun {
		context.Log.Debugf("IsReady() dry run for %s", h.HelmRelease.Name)
		return true, ctrl.Result{}, nil
	}

	deployed, err := helm.IsReleaseDeployed(h.HelmRelease.Name, h.HelmRelease.Namespace)
	if err != nil {
		context.Log.ErrorfThrottled("Error occurred checking release deloyment: %v", err.Error())
		return false, ctrl.Result{}, err
	}

	return !deployed, ctrl.Result{}, nil
}

// PostActionUpdateStatus does installation post-action
func (h HelmHandler) PostActionUpdateStatus(ctx actionspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// PostAction does uninstall post-action
func (h HelmHandler) PostAction(context actionspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// IsPostActionDone returns true if post-action done
func (h HelmHandler) IsPostActionDone(context actionspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// CompletedActionUpdateStatus does the lifecycle completed Action status update
func (h HelmHandler) CompletedActionUpdateStatus(ctx actionspi.HandlerContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateStatus(ctx, moduleapi.CondInstallComplete, moduleapi.StateCompleted)
}
