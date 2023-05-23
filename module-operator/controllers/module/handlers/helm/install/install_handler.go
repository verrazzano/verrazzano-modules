// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package install

import (
	"context"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/handlers/common"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/handlerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/util"
	helm2 "github.com/verrazzano/verrazzano-modules/pkg/helm"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	"helm.sh/helm/v3/pkg/release"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

type HelmHandler struct {
	common.BaseHandler
}

// upgradeFuncSig is a function needed for unit test override
type upgradeFuncSig func(log vzlog.VerrazzanoLogger, releaseOpts *helm2.HelmReleaseOpts, wait bool, dryRun bool) (*release.Release, error)

var (
	_ handlerspi.StateMachineHandler = &HelmHandler{}

	upgradeFunc upgradeFuncSig = helm2.UpgradeRelease
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
	return "install"
}

// IsWorkNeeded returns true if install is needed
func (h HelmHandler) IsWorkNeeded(ctx handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// PreWorkUpdateStatus does the pre-Work status update
func (h HelmHandler) PreWorkUpdateStatus(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return h.UpdateStatus(ctx, moduleapi.CondPreInstall, moduleapi.ModuleStateReconciling)
}

// PreWork does the pre-work
func (h HelmHandler) PreWork(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	// Update the spec version if it is not set
	if len(h.ModuleCR.Spec.Version) == 0 {
		// Update spec version to match chart, always requeue to get ModuleCR with version
		h.ModuleCR.Spec.Version = h.Config.ChartInfo.Version
		if err := ctx.Client.Update(context.TODO(), h.ModuleCR); err != nil {
			return util.NewRequeueWithShortDelay(), nil
		}
		// ALways reconcile so that we get a new tracker with the latest ModuleCR
		return util.NewRequeueWithDelay(1, 2, time.Second), nil
	}

	return ctrl.Result{}, nil
}

// DoWorkUpdateStatus does th status update
func (h HelmHandler) DoWorkUpdateStatus(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return h.UpdateStatus(ctx, moduleapi.CondInstallStarted, moduleapi.ModuleStateReconciling)
}

// DoWork installs the module using Helm
func (h HelmHandler) DoWork(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	installed, err := helm2.IsReleaseInstalled(h.HelmRelease.Name, h.HelmRelease.Namespace)
	if err != nil {
		ctx.Log.ErrorfThrottled("Error checking if Helm release installed for %s/%s", h.HelmRelease.Namespace, h.HelmRelease.Name)
		return ctrl.Result{}, err
	}
	if installed {
		return ctrl.Result{}, nil
	}

	// Perform a Helm install using the helm upgrade --install command
	helmRelease := h.BaseHandler.Config.HelmInfo.HelmRelease
	helmOverrides, err := helm2.LoadOverrideFiles(ctx.Log, ctx.Client, helmRelease.Name, h.ModuleCR.Namespace, h.ModuleCR.Spec.Overrides)
	if err != nil {
		return ctrl.Result{}, err
	}
	var opts = &helm2.HelmReleaseOpts{
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

// IsWorkDone Indicates whether a module is installed and ready
func (h HelmHandler) IsWorkDone(ctx handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	if ctx.DryRun {
		ctx.Log.Debugf("IsReady() dry run for %s", h.BaseHandler.Name)
		return true, ctrl.Result{}, nil
	}

	deployed, err := helm2.IsReleaseDeployed(h.BaseHandler.Name, h.BaseHandler.Namespace)
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

// PostWorkUpdateStatus does the post-work status update
func (h HelmHandler) PostWorkUpdateStatus(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// PostWork does installation pre-work
func (h HelmHandler) PostWork(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// WorkCompletedUpdateStatus updates the status to completed
func (h HelmHandler) WorkCompletedUpdateStatus(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateDoneStatus(ctx, moduleapi.CondInstallComplete, moduleapi.ModuleStateReady, h.ModuleCR.Spec.Version)
}
