// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package common

import (
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/handlerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/constants"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/util"
	helm2 "github.com/verrazzano/verrazzano-modules/pkg/helm"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	"helm.sh/helm/v3/pkg/release"
	ctrl "sigs.k8s.io/controller-runtime"
)

// readyConditionMessages defines the condition messages for the Ready type condition
var readyConditionMessages = map[moduleapi.ModuleConditionReason]string{
	moduleapi.ReadyReasonInstallStarted:     "Started installing Module %s as Helm release %s/%s",
	moduleapi.ReadyReasonInstallSucceeded:   "Successfully installed Module %s as Helm release %s/%s",
	moduleapi.ReadyReasonInstallFailed:      "Failed installing Module %s as Helm release %s%s: %v",
	moduleapi.ReadyReasonUninstallStarted:   "Started uninstalling Module %s as Helm release %s/%s",
	moduleapi.ReadyReasonUninstallSucceeded: "Successfully uninstalled Module %s as Helm release %s/%s",
	moduleapi.ReadyReasonUninstallFailed:    "Failed uninstalling Module %s as Helm release %s/%s: %v",
	moduleapi.ReadyReasonUpdateStarted:      "Started updating Module %s as Helm release %s/%s",
	moduleapi.ReadyReasonUpdateSucceeded:    "Successfully updated Module %s as Helm release %s/%s",
	moduleapi.ReadyReasonUpdateFailed:       "Failed updating Module %s as Helm release %s/%s: %v",
	moduleapi.ReadyReasonUpgradeStarted:     "Started upgrading Module %s as Helm release %s/%s",
	moduleapi.ReadyReasonUpgradeSucceeded:   "Successfully upgraded Module %s as Helm release %s/%s",
	moduleapi.ReadyReasonUpgradeFailed:      "Failed upgrading Module %s as Helm release %s/%s: %v",
}

// upgradeFuncSig is a function needed for unit test override
type upgradeFuncSig func(log vzlog.VerrazzanoLogger, releaseOpts *helm2.HelmReleaseOpts, wait bool, dryRun bool) (*release.Release, error)

var upgradeFunc upgradeFuncSig = helm2.UpgradeRelease

type BaseHandler struct {
	// Config is the handler configuration
	Config handlerspi.StateMachineHandlerConfig

	// HelmInfo has the helm information
	handlerspi.HelmInfo

	// ModuleCR is the Module CR being handled
	ModuleCR *moduleapi.Module

	// ImagePullSecretKeyname is the Helm Value Key for the image pull secret for a chart
	ImagePullSecretKeyname string
}

func SetUpgradeFunc(f upgradeFuncSig) {
	upgradeFunc = f
}

func ResetUpgradeFunc() {
	upgradeFunc = helm2.UpgradeRelease
}

// Init initializes the handler with Helm chart information
func (h *BaseHandler) InitHandler(_ handlerspi.HandlerContext, config handlerspi.StateMachineHandlerConfig) (ctrl.Result, error) {
	h.Config = config
	h.HelmInfo = config.HelmInfo
	h.ImagePullSecretKeyname = constants.GlobalImagePullSecName
	h.ModuleCR = config.CR.(*moduleapi.Module)
	return ctrl.Result{}, nil
}

// HelmUpgradeOrInstall does a Helm upgrade --install of the chart
func (h BaseHandler) HelmUpgradeOrInstall(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	// Perform a Helm install using the helm upgrade --install command
	helmRelease := h.Config.HelmInfo.HelmRelease
	helmOverrides, err := helm2.LoadOverrideFiles(ctx.Log, ctx.Client, helmRelease.Name, h.ModuleCR.Namespace, h.ModuleCR.Spec.Overrides)
	if err != nil {
		return ctrl.Result{}, err
	}
	var opts = &helm2.HelmReleaseOpts{
		RepoURL:      helmRelease.Repository.URI,
		ReleaseName:  helmRelease.Name,
		Namespace:    helmRelease.Namespace,
		ChartPath:    helmRelease.ChartInfo.Path,
		ChartVersion: helmRelease.ChartInfo.Version,
		Overrides:    helmOverrides,
	}
	_, err = upgradeFunc(ctx.Log, opts, false, ctx.DryRun)
	return ctrl.Result{}, err
}

// CheckReleaseDeployedAndReady checks if the Helm release is deployed and ready
func (h BaseHandler) CheckReleaseDeployedAndReady(ctx handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	if ctx.DryRun {
		ctx.Log.Debugf("IsReady() dry run for %s", h.HelmRelease.Name)
		return true, ctrl.Result{}, nil
	}
	// Check if the Helm release is deployed
	deployed, err := helm2.IsReleaseDeployed(h.HelmRelease.Name, h.HelmRelease.Namespace)
	if err != nil {
		ctx.Log.ErrorfThrottled("Error occurred checking release deployment: %v", err.Error())
		return false, ctrl.Result{}, err
	}
	if !deployed {
		return false, util.NewRequeueWithShortDelay(), nil
	}

	// Check if the workload pods are ready
	ready, err := CheckWorkLoadsReady(ctx, h.HelmRelease.Name, h.HelmRelease.Namespace)
	return ready, ctrl.Result{}, err
}
