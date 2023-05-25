// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package common

import (
	"context"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/handlerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/constants"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/util"
	helm2 "github.com/verrazzano/verrazzano-modules/pkg/helm"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	"helm.sh/helm/v3/pkg/release"
	ctrl "sigs.k8s.io/controller-runtime"
)

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

// Init initializes the handler with Helm chart information
func (h *BaseHandler) InitHandler(_ handlerspi.HandlerContext, config handlerspi.StateMachineHandlerConfig) (ctrl.Result, error) {
	h.Config = config
	h.HelmInfo = config.HelmInfo
	h.ImagePullSecretKeyname = constants.GlobalImagePullSecName
	h.ModuleCR = config.CR.(*moduleapi.Module)
	return ctrl.Result{}, nil
}

// UpdateStatus does the lifecycle pre-Work status update
func (h BaseHandler) UpdateStatus(ctx handlerspi.HandlerContext, cond moduleapi.LifecycleCondition, state moduleapi.ModuleStateType) (ctrl.Result, error) {
	AppendCondition(h.ModuleCR, string(cond), cond)
	h.ModuleCR.Status.State = state
	if err := ctx.Client.Status().Update(context.TODO(), h.ModuleCR); err != nil {
		return util.NewRequeueWithShortDelay(), nil
	}
	return ctrl.Result{}, nil
}

// UpdateDoneStatus does the lifecycle status update when the work is done
func (h BaseHandler) UpdateDoneStatus(ctx handlerspi.HandlerContext, cond moduleapi.LifecycleCondition, state moduleapi.ModuleStateType, version string) (ctrl.Result, error) {
	AppendCondition(h.ModuleCR, string(cond), cond)
	h.ModuleCR.Status.State = state
	h.ModuleCR.Status.ObservedGeneration = h.ModuleCR.Generation
	if len(version) > 0 {
		h.ModuleCR.Status.Version = version
	}
	if err := ctx.Client.Status().Update(context.TODO(), h.ModuleCR); err != nil {
		return util.NewRequeueWithShortDelay(), nil
	}
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
		// TODO -- pull from a secret ref?
		//Username:     "",
		//Password:     "",
	}
	_, err = upgradeFunc(ctx.Log, opts, true, ctx.DryRun)
	return ctrl.Result{}, err
}

// CheckReleaseDeployedAndReady checks if the Helm release is deployed and ready
func (h BaseHandler) CheckReleaseDeployedAndReady(ctx handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	if ctx.DryRun {
		ctx.Log.Debugf("IsReady() dry run for %s", h.HelmRelease.Name)
		return true, ctrl.Result{}, nil
	}

	deployed, err := helm2.IsReleaseDeployed(h.HelmRelease.Name, h.HelmRelease.Namespace)
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
