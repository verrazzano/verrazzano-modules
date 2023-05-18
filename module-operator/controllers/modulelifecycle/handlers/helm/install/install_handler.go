// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package install

import (
	"github.com/verrazzano/verrazzano-modules/common/handlerspi"
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
	_ handlerspi.StateMachineHandler = &HelmHandler{}

	upgradeFunc upgradeFuncSig = helm.UpgradeRelease
)

func NewHandler() handlerspi.StateMachineHandler {
	return &HelmHandler{}
}

// Init initializes the handler with Helm chart information
func (h *HelmHandler) Init(ctx handlerspi.HandlerContext, config handlerspi.StateMachineHandlerConfig) (ctrl.Result, error) {
	return h.BaseHandler.Init(ctx, config)
}

// GetActionName returns the action name
func (h HelmHandler) GetActionName() string {
	return "install"
}

// IsActionNeeded returns true if install is needed
func (h HelmHandler) IsActionNeeded(ctx handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	if h.ModuleCR.Status.State == moduleapi.StateCompleted || h.ModuleCR.Status.State == moduleapi.StateNotNeeded {
		return false, ctrl.Result{}, nil
	}
	return true, ctrl.Result{}, nil
}

// PreActionUpdateStatus does the lifecycle pre-Action status update
func (h HelmHandler) PreActionUpdateStatus(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return h.UpdateStatus(ctx, moduleapi.CondPreInstall, moduleapi.ModuleStateReconciling)
}

// PreAction does installation pre-action
func (h HelmHandler) PreAction(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// IsPreActionDone returns true if pre-action done
func (h HelmHandler) IsPreActionDone(ctx handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// ActionUpdateStatus does the lifecycle Action status update
func (h HelmHandler) ActionUpdateStatus(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return h.UpdateStatus(ctx, moduleapi.CondInstallStarted, moduleapi.ModuleStateReconciling)
}

// DoAction installs the module using Helm
func (h HelmHandler) DoAction(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	installed, err := helm.IsReleaseInstalled(h.HelmRelease.Name, h.HelmRelease.Namespace)
	if err != nil {
		ctx.Log.ErrorfThrottled("Error checking if Helm release installed for %s/%s", h.HelmRelease.Namespace, h.HelmRelease.Name)
		return ctrl.Result{}, err
	}
	if installed {
		return ctrl.Result{}, nil
	}

	// Perform a Helm install using the helm upgrade --install command
	helmRelease := h.BaseHandler.Config.HelmInfo.HelmRelease
	helmOverrides, err := helm.LoadOverrideFiles(ctx.Log, ctx.Client, helmRelease.Name, h.BaseHandler.ModuleCR.Namespace, helmRelease.Overrides)
	if err != nil {
		return ctrl.Result{}, err
	}
	var opts = &helm.HelmReleaseOpts{
		RepoURL:      helmRelease.Repository.URI,
		ReleaseName:  h.BaseHandler.Name,
		Namespace:    h.BaseHandler.Namespace,
		ChartPath:    helmRelease.ChartInfo.Path,
		ChartVersion: helmRelease.ChartInfo.Version,
		Overrides:    helmOverrides,
		// TODO -- pull from a secret ref?
		//Username:     "",
		//Password:     "",
	}
	_, err = upgradeFunc(ctx.Log, opts, false, ctx.DryRun)
	return ctrl.Result{}, err
}

// IsActionDone Indicates whether a module is installed and ready
func (h HelmHandler) IsActionDone(ctx handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	if ctx.DryRun {
		ctx.Log.Debugf("IsReady() dry run for %s", h.BaseHandler.Name)
		return true, ctrl.Result{}, nil
	}

	deployed, err := helm.IsReleaseDeployed(h.BaseHandler.Name, h.BaseHandler.Namespace)
	if err != nil {
		ctx.Log.ErrorfThrottled("Error occurred checking release deployment: %v", err.Error())
		return false, ctrl.Result{}, err
	}
	if !deployed {
		return false, util.NewRequeueWithShortDelay(), nil
	}
	// check if helm release at the correct version
	if !h.releaseVersionMatches(ctx.Log) {
		return false, util.NewRequeueWithShortDelay(), nil
	}

	// TODO check if release is ready (check deployments)
	return true, ctrl.Result{}, err
}

// PostActionUpdateStatus does installation post-action
func (h HelmHandler) PostActionUpdateStatus(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// PostAction does installation pre-action
func (h HelmHandler) PostAction(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// IsPostActionDone returns true if post-action done
func (h HelmHandler) IsPostActionDone(ctx handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// CompletedActionUpdateStatus does the lifecycle completed Action status update
func (h HelmHandler) CompletedActionUpdateStatus(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateStatus(ctx, moduleapi.CondInstallComplete, moduleapi.StateCompleted)
}
