// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package update

import (
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/handlers/common"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/status"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/result"
	handlerspi2 "github.com/verrazzano/verrazzano-modules/pkg/controller/spi/handlerspi"
)

type HelmHandler struct {
	common.BaseHandler
}

var (
	_ handlerspi2.StateMachineHandler = &HelmHandler{}
)

func NewHandler() handlerspi2.StateMachineHandler {
	return &HelmHandler{}
}

// GetWorkName returns the work name
func (h HelmHandler) GetWorkName() string {
	return "update"
}

// IsWorkNeeded returns true if update is needed
func (h HelmHandler) IsWorkNeeded(ctx handlerspi2.HandlerContext) (bool, result.Result) {
	return true, result.NewResult()
}

// PreWorkUpdateStatus updates the status for the pre-work state
func (h HelmHandler) PreWorkUpdateStatus(ctx handlerspi2.HandlerContext) result.Result {
	return result.NewResult()
}

// PreWork does the pre-work
func (h HelmHandler) PreWork(ctx handlerspi2.HandlerContext) result.Result {
	return result.NewResult()
}

// DoWorkUpdateStatus updates the status for the work state
func (h HelmHandler) DoWorkUpdateStatus(ctx handlerspi2.HandlerContext) result.Result {
	module := ctx.CR.(*moduleapi.Module)
	return status.UpdateReadyConditionReconciling(ctx, module, moduleapi.ReadyReasonUpdateStarted)
}

// DoWork updates the module using Helm
func (h HelmHandler) DoWork(ctx handlerspi2.HandlerContext) result.Result {
	return h.HelmUpgradeOrInstall(ctx)
}

// IsWorkDone Indicates whether a module is updated and ready
func (h HelmHandler) IsWorkDone(ctx handlerspi2.HandlerContext) (bool, result.Result) {
	return h.CheckReleaseDeployedAndReady(ctx)
}

// PostWorkUpdateStatus does the post-work status update
func (h HelmHandler) PostWorkUpdateStatus(ctx handlerspi2.HandlerContext) result.Result {
	return result.NewResult()
}

// PostWork does installation pre-work
func (h HelmHandler) PostWork(ctx handlerspi2.HandlerContext) result.Result {
	return result.NewResult()
}

// WorkCompletedUpdateStatus updates the status to completed
func (h HelmHandler) WorkCompletedUpdateStatus(ctx handlerspi2.HandlerContext) result.Result {
	module := ctx.CR.(*moduleapi.Module)
	return status.UpdateReadyConditionSucceeded(ctx, module, moduleapi.ReadyReasonUpdateSucceeded)
}
