// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package upgrade

import (
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/handlers/common"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/status"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/result"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/spi/handlerspi"
	helm2 "github.com/verrazzano/verrazzano-modules/pkg/helm"
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

// GetWorkName returns the work name
func (h HelmHandler) GetWorkName() string {
	return "upgrade"
}

// IsWorkNeeded returns true if upgrade is needed
func (h HelmHandler) IsWorkNeeded(ctx handlerspi.HandlerContext) (bool, result.Result) {
	module := ctx.CR.(*moduleapi.Module)

	installed, err := helm2.IsReleaseInstalled(ctx.HelmRelease.Name, module.Spec.TargetNamespace)
	if err != nil {
		ctx.Log.ErrorfThrottled("Error checking if Helm release installed for %s/%s", ctx.HelmRelease.Namespace, ctx.HelmRelease.Name)
		return true, result.NewResult()
	}
	return installed, result.NewResult()
}

// PreWorkUpdateStatus updates the status for the pre-work state
func (h HelmHandler) PreWorkUpdateStatus(ctx handlerspi.HandlerContext) result.Result {
	return result.NewResult()
}

// PreWork does the pre-work
func (h HelmHandler) PreWork(ctx handlerspi.HandlerContext) result.Result {
	return result.NewResult()
}

func (h HelmHandler)CheckDependencies(context handlerspi.HandlerContext) result.Result {
	return result.NewResult()
}

// DoWorkUpdateStatus updates the status for the work state
func (h HelmHandler) DoWorkUpdateStatus(ctx handlerspi.HandlerContext) result.Result {
	module := ctx.CR.(*moduleapi.Module)
	return status.UpdateReadyConditionReconciling(ctx, module, moduleapi.ReadyReasonUpgradeStarted)
}

// DoWork upgrades the module using Helm
func (h HelmHandler) DoWork(ctx handlerspi.HandlerContext) result.Result {
	return h.HelmUpgradeOrInstall(ctx)
}

// IsWorkDone indicates whether a module is upgraded and ready
func (h HelmHandler) IsWorkDone(ctx handlerspi.HandlerContext) (bool, result.Result) {
	return h.CheckReleaseDeployedAndReady(ctx)
}

// PostWorkUpdateStatus does the post-work status update
func (h HelmHandler) PostWorkUpdateStatus(ctx handlerspi.HandlerContext) result.Result {
	return result.NewResult()
}

// PostWork does installation pre-work
func (h HelmHandler) PostWork(ctx handlerspi.HandlerContext) result.Result {
	return result.NewResult()
}

// WorkCompletedUpdateStatus updates the status to completed
func (h HelmHandler) WorkCompletedUpdateStatus(ctx handlerspi.HandlerContext) result.Result {
	module := ctx.CR.(*moduleapi.Module)
	return status.UpdateReadyConditionSucceeded(ctx, module, moduleapi.ReadyReasonUpgradeSucceeded)
}
