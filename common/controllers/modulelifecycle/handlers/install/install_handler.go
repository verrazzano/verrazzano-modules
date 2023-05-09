// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package install

import (
	actionspi "github.com/verrazzano/verrazzano-modules/common/actionspi"
	"github.com/verrazzano/verrazzano-modules/common/controllers/modulelifecycle/handlers/common"
	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	"github.com/verrazzano/verrazzano-modules/common/pkg/helm"
	"github.com/verrazzano/verrazzano-modules/common/pkg/vzlog"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	vzhelm "github.com/verrazzano/verrazzano/pkg/helm"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	"helm.sh/helm/v3/pkg/release"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Handler struct {
	common.BaseHandler
}

// upgradeFuncSig is a function needed for unit test override
type upgradeFuncSig func(log vzlog.VerrazzanoLogger, releaseOpts *helm.HelmReleaseOpts, wait bool, dryRun bool) (*release.Release, error)

var (
	_ actionspi.LifecycleActionHandler = &Handler{}

	upgradeFunc upgradeFuncSig = helm.UpgradeRelease
)

func NewComponent() actionspi.LifecycleActionHandler {
	return &Handler{}
}

// Init initializes the component with Helm chart information
func (h *Handler) Init(ctx spi.ComponentContext, config actionspi.HandlerConfig) (ctrl.Result, error) {
	return h.BaseHandler.Init(ctx, config)
}

// GetActionName returns the action name
func (h Handler) GetActionName() string {
	return "install"
}

// IsActionNeeded returns true if install is needed
func (h Handler) IsActionNeeded(ctx spi.ComponentContext) (bool, ctrl.Result, error) {
	if h.ModuleCR.Status.State == moduleapi.StateCompleted || h.BaseHandler.ModuleCR.Status.State == moduleapi.StateNotNeeded {
		return false, ctrl.Result{}, nil
	}
	return true, ctrl.Result{}, nil
}

// PreActionUpdateStatus does the lifecycle pre-Action status update
func (h Handler) PreActionUpdateStatus(ctx spi.ComponentContext) (ctrl.Result, error) {
	return h.UpdateStatus(ctx, moduleapi.CondPreInstall, moduleapi.ModuleStateReconciling)
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
	return h.UpdateStatus(ctx, moduleapi.CondInstallStarted, moduleapi.ModuleStateReconciling)
}

// DoAction installs the component using Helm
func (h Handler) DoAction(ctx spi.ComponentContext) (ctrl.Result, error) {
	installed, err := vzhelm.IsReleaseInstalled(h.BaseHandler.Name, h.BaseHandler.Namespace)
	if err != nil {
		ctx.Log().ErrorfThrottled("Error checking if Helm release installed for %s/%s", h.BaseHandler.Config.ChartDir, h.BaseHandler.Name)
		return ctrl.Result{}, err
	}
	if installed {
		return ctrl.Result{}, nil
	}

	// Perform a Helm install using the helm upgrade --install command
	helmRelease := h.BaseHandler.Config.HelmInfo.HelmRelease
	helmOverrides, err := helm.LoadOverrideFiles(ctx, helmRelease.Name, h.BaseHandler.ModuleCR.Namespace, helmRelease.Overrides)
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
	_, err = upgradeFunc(ctx.Log(), opts, h.BaseHandler.WaitForInstall, ctx.IsDryRun())
	return ctrl.Result{}, err
}

// IsActionDone Indicates whether a component is installed and ready
func (h Handler) IsActionDone(ctx spi.ComponentContext) (bool, ctrl.Result, error) {
	if ctx.IsDryRun() {
		ctx.Log().Debugf("IsReady() dry run for %s", h.BaseHandler.ReleaseName)
		return true, ctrl.Result{}, nil
	}

	deployed, err := vzhelm.IsReleaseDeployed(h.BaseHandler.ReleaseName, h.BaseHandler.ChartNamespace)
	if err != nil {
		ctx.Log().ErrorfThrottled("Error occurred checking release deployment: %v", err.Error())
		return false, ctrl.Result{}, err
	}
	if !deployed {
		return false, util.NewRequeueWithShortDelay(), nil
	}
	// check if helm release at the correct version
	if !h.releaseVersionMatches(ctx.Log()) {
		return false, util.NewRequeueWithShortDelay(), nil
	}

	// TODO check if release is ready (check deployments)
	return true, ctrl.Result{}, err
}

// PostActionUpdateStatus does installation post-action
func (h Handler) PostActionUpdateStatus(ctx spi.ComponentContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// PostAction does installation pre-action
func (h Handler) PostAction(ctx spi.ComponentContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// IsPostActionDone returns true if post-action done
func (h Handler) IsPostActionDone(ctx spi.ComponentContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// CompletedActionUpdateStatus does the lifecycle completed Action status update
func (h Handler) CompletedActionUpdateStatus(ctx spi.ComponentContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateStatus(ctx, moduleapi.CondInstallComplete, moduleapi.StateCompleted)
}
