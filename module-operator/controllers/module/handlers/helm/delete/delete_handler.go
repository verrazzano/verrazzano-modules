// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package delete

import (
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/handlers/common"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/handlerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/helm"
	ctrl "sigs.k8s.io/controller-runtime"
)

type HelmHandler struct {
	common.BaseHandler
}

var (
	_ handlerspi.StateMachineHandler = &HelmHandler{}
)

func NewHandler() handlerspi.StateMachineHandler {
	return &HelmHandler{}
}

// Init initializes the handler with Helm chart information
func (h *HelmHandler) Init(ctx handlerspi.HandlerContext, config handlerspi.StateMachineHandlerConfig) (ctrl.Result, error) {
	return h.BaseHandler.InitHandler(ctx, config)
}

// GetWorkName returns the work name
func (h HelmHandler) GetWorkName() string {
	return "uninstall"
}

// IsWorkNeeded returns true if install is needed
func (h HelmHandler) IsWorkNeeded(ctx handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// PreWorkUpdateStatus does the lifecycle pre-Work status update
func (h HelmHandler) PreWorkUpdateStatus(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// PreWork does the pre-work
func (h HelmHandler) PreWork(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// DoWorkUpdateStatus does the work status update
func (h HelmHandler) DoWorkUpdateStatus(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateStatus(ctx, moduleapi.ReadyReasonUninstallStarted)
}

// DoWork uninstalls the module using Helm
func (h HelmHandler) DoWork(context handlerspi.HandlerContext) (ctrl.Result, error) {
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

// IsWorkDone Indicates whether a module is uninstalled
func (h HelmHandler) IsWorkDone(context handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	if context.DryRun {
		context.Log.Debugf("IsReady() dry run for %s", h.HelmRelease.Name)
		return true, ctrl.Result{}, nil
	}

	deployed, err := helm.IsReleaseDeployed(h.HelmRelease.Name, h.HelmRelease.Namespace)
	if err != nil {
		context.Log.ErrorfThrottled("Error occurred checking release deployment: %v", err.Error())
		return false, ctrl.Result{}, err
	}

	return !deployed, ctrl.Result{}, nil
}

// PostWorkUpdateStatus does the post-work status update
func (h HelmHandler) PostWorkUpdateStatus(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// PostWork does installation pre-work
func (h HelmHandler) PostWork(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// CompletedWorkUpdateStatus does the lifecycle completed Work status update
func (h HelmHandler) WorkCompletedUpdateStatus(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateStatus(ctx, moduleapi.ReadyReasonUnInstallSucceeded)
}
