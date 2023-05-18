// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package upgrade

import (
	actionspi "github.com/verrazzano/verrazzano-modules/common/handlerspi"
	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	"github.com/verrazzano/verrazzano-modules/common/pkg/helm"
	"github.com/verrazzano/verrazzano-modules/common/pkg/vzlog"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/modulelifecycle/handlers/common"
	"helm.sh/helm/v3/pkg/release"
	ctrl "sigs.k8s.io/controller-runtime"
)

type HelmHandler struct {
	common.BaseHandler
}

// upgradeFuncSig is a function needed for unit test override
type upgradeFuncSig func(log vzlog.VerrazzanoLogger, releaseOpts *helm.HelmReleaseOpts, wait bool, dryRun bool) (*release.Release, error)

var (
	_ actionspi.StateMachineHandler = &HelmHandler{}

	upgradeFunc upgradeFuncSig = helm.UpgradeRelease
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
	return "install"
}

// IsActionNeeded returns true if install is needed
func (h HelmHandler) IsActionNeeded(ctx actionspi.HandlerContext) (bool, ctrl.Result, error) {
	installed, err := helm.IsReleaseInstalled(h.HelmRelease.Name, h.BaseHandler.Config.Namespace)
	if err != nil {
		ctx.Log.ErrorfThrottled("Error checking if Helm release installed for %s/%s", h.HelmRelease.Namespace, h.HelmRelease.Name)
		return true, ctrl.Result{}, err
	}
	return installed, ctrl.Result{}, err
}

// PreActionUpdateStatus does the lifecycle pre-Action status update
func (h HelmHandler) PreActionUpdateStatus(ctx actionspi.HandlerContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateStatus(ctx, moduleapi.CondPreUpgrade, moduleapi.ModuleStateReconciling)
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
	return h.BaseHandler.UpdateStatus(ctx, moduleapi.CondUpgradeStarted, moduleapi.ModuleStateReconciling)
}

// DoAction installs the module using Helm
func (h HelmHandler) DoAction(ctx actionspi.HandlerContext) (ctrl.Result, error) {
	// Perform a Helm install using the helm upgrade --install command
	helmRelease := h.BaseHandler.Config.HelmInfo.HelmRelease
	helmOverrides, err := helm.LoadOverrideFiles(ctx.Log, ctx.Client, helmRelease.Name, h.BaseHandler.ModuleCR.Namespace, helmRelease.Overrides)
	if err != nil {
		return ctrl.Result{}, err
	}
	var opts = &helm.HelmReleaseOpts{
		RepoURL:      helmRelease.Repository.URI,
		ReleaseName:  helmRelease.Name,
		Namespace:    helmRelease.Namespace,
		ChartPath:    helmRelease.ChartInfo.Path,
		ChartVersion: helmRelease.ChartInfo.Version,
		Overrides:    helmOverrides,
		// TODO -- pull from a secret ref?
		//Username:     "",
		//Password:     "",
	}
	_, err = upgradeFunc(ctx.Log, opts, true, ctx.DryRun)
	return ctrl.Result{}, err
}

// IsActionDone Indicates whether a module is installed and ready
func (h HelmHandler) IsActionDone(ctx actionspi.HandlerContext) (bool, ctrl.Result, error) {
	if ctx.DryRun {
		ctx.Log.Debugf("IsReady() dry run for %s", h.HelmRelease.Name)
		return true, ctrl.Result{}, nil
	}

	deployed, err := helm.IsReleaseDeployed(h.HelmRelease.Name, h.HelmRelease.Namespace)
	if err != nil {
		ctx.Log.ErrorfThrottled("Error occurred checking release deployment: %v", err.Error())
		return false, ctrl.Result{}, err
	}
	if !deployed {
		return false, util.NewRequeueWithShortDelay(), nil
	}

	// TODO check if release is ready (check deployments)
	return true, ctrl.Result{}, err
}

// PostActionUpdateStatue does installation post-action
func (h HelmHandler) PostActionUpdateStatus(ctx actionspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// PostAction does installation pre-action
func (h HelmHandler) PostAction(ctx actionspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// IsPostActionDone returns true if post-action done
func (h HelmHandler) IsPostActionDone(ctx actionspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// CompletedActionUpdateStatus does the lifecycle completed Action status update
func (h HelmHandler) CompletedActionUpdateStatus(ctx actionspi.HandlerContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateStatus(ctx, moduleapi.CondUpgradeComplete, moduleapi.StateCompleted)
}
