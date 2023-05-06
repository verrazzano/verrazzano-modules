// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package uninstall

import (
	actionspi "github.com/verrazzano/verrazzano-modules/common/actionspi"
	"github.com/verrazzano/verrazzano-modules/common/controllers/modulelifecycle/handlers/common"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	vzhelm "github.com/verrazzano/verrazzano/pkg/helm"
	"github.com/verrazzano/verrazzano/platform-operator/constants"
	helmcomp "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/helm"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Handler struct {
	BaseHandler common.BaseHandler
}

var (
	_ actionspi.LifecycleActionHandler = &Handler{}
)

func NewComponent() actionspi.LifecycleActionHandler {
	return &Handler{}
}

// Init initializes the component with Helm chart information
func (h *Handler) Init(_ spi.ComponentContext, config actionspi.HandlerConfig) (ctrl.Result, error) {
	h.BaseHandler.HelmComponent = helmcomp.HelmComponent{
		ReleaseName:             config.HelmInfo.HelmRelease.Name,
		ChartNamespace:          config.HelmInfo.HelmRelease.Namespace,
		ChartDir:                config.ChartDir,
		IgnoreNamespaceOverride: true,
		ImagePullSecretKeyname:  constants.GlobalImagePullSecName,
	}
	h.BaseHandler.CR = config.CR.(*moduleapi.ModuleLifecycle)
	h.BaseHandler.Config = config
	return ctrl.Result{}, nil
}

// GetActionName returns the action name
func (h Handler) GetActionName() string {
	return "uninstall"
}

// IsActionNeeded returns true if uninstall is needed
func (h Handler) IsActionNeeded(context spi.ComponentContext) (bool, ctrl.Result, error) {
	installed, err := vzhelm.IsReleaseInstalled(h.BaseHandler.ReleaseName, h.BaseHandler.Config.Namespace)
	if err != nil {
		context.Log().ErrorfThrottled("Error checking if Helm release installed for %s/%s", h.BaseHandler.Config.ChartDir, h.BaseHandler.ReleaseName)
		return true, ctrl.Result{}, err
	}
	return installed, ctrl.Result{}, err
}

// PreActionUpdateStatus does the lifecycle pre-Action status update
func (h Handler) PreActionUpdateStatus(ctx spi.ComponentContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateStatus(ctx, moduleapi.CondPreUninstall, moduleapi.ModuleStateReconciling)
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
	return h.BaseHandler.UpdateStatus(ctx, moduleapi.CondUninstallStarted, moduleapi.ModuleStateReconciling)
}

// DoAction uninstalls the component using Helm
func (h Handler) DoAction(context spi.ComponentContext) (ctrl.Result, error) {
	installed, err := vzhelm.IsReleaseInstalled(h.BaseHandler.ReleaseName, h.BaseHandler.Config.Namespace)
	if err != nil {
		context.Log().ErrorfThrottled("Error checking if Helm release installed for %s/%s", h.BaseHandler.Config.ChartDir, h.BaseHandler.ReleaseName)
		return ctrl.Result{}, err
	}
	if !installed {
		return ctrl.Result{}, err
	}

	err = vzhelm.Uninstall(context.Log(), h.BaseHandler.ReleaseName, h.BaseHandler.ChartNamespace, context.IsDryRun())
	return ctrl.Result{}, err
}

// IsActionDone Indicates whether a component is uninstalled
func (h Handler) IsActionDone(context spi.ComponentContext) (bool, ctrl.Result, error) {
	if context.IsDryRun() {
		context.Log().Debugf("IsReady() dry run for %s", h.BaseHandler.ReleaseName)
		return true, ctrl.Result{}, nil
	}

	deployed, err := vzhelm.IsReleaseDeployed(h.BaseHandler.ReleaseName, h.BaseHandler.ChartNamespace)
	if err != nil {
		context.Log().ErrorfThrottled("Error occurred checking release deloyment: %v", err.Error())
		return false, ctrl.Result{}, err
	}

	return !deployed, ctrl.Result{}, nil
}

// PostActionUpdateStatus does installation post-action
func (h Handler) PostActionUpdateStatus(ctx spi.ComponentContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// PostAction does uninstall post-action
func (h Handler) PostAction(context spi.ComponentContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// IsPostActionDone returns true if post-action done
func (h Handler) IsPostActionDone(context spi.ComponentContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// CompletedActionUpdateStatus does the lifecycle completed Action status update
func (h Handler) CompletedActionUpdateStatus(ctx spi.ComponentContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateStatus(ctx, moduleapi.CondInstallComplete, moduleapi.StateCompleted)
}
