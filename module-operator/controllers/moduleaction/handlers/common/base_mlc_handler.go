// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package common

import (
	"context"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/handlerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/constants"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/util"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ handlerspi.StateMachineHandler = &BaseHandler{}

type BaseHandler struct {
	// Config is the handler configuration
	Config handlerspi.StateMachineHandlerConfig

	// HelmInfo has the helm information
	handlerspi.HelmInfo

	// ModuleCR is the ModuleAction CR being handled
	ModuleCR *moduleapi.ModuleAction

	// ChartDir is the helm chart directory (TODO remove this and use HelmInfo path)
	ChartDir string

	// ImagePullSecretKeyname is the Helm Value Key for the image pull secret for a chart
	ImagePullSecretKeyname string
}

func (h *BaseHandler) GetWorkName() string {
	//TODO implement me
	return "unknown"
}

// Init initializes the handler with Helm chart information
func (h *BaseHandler) Init(_ handlerspi.HandlerContext, config handlerspi.StateMachineHandlerConfig) (ctrl.Result, error) {
	h.Config = config
	h.HelmInfo = config.HelmInfo
	h.ImagePullSecretKeyname = constants.GlobalImagePullSecName
	h.ModuleCR = config.CR.(*moduleapi.ModuleAction)
	return ctrl.Result{}, nil
}

// IsWorkNeeded returns true if install is needed
func (h BaseHandler) IsWorkNeeded(ctx handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// PreWorkUpdateStatus does the lifecycle pre-Work status update
func (h *BaseHandler) PreWorkUpdateStatus(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// PreWork does the pre-action
func (h BaseHandler) PreWork(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// WorkUpdateStatus does the lifecycle Work status update
func (h *BaseHandler) DoWorkUpdateStatus(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// DoWork installs the module using Helm
func (h *BaseHandler) DoWork(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// IsWorkDone Indicates whether a module is installed and ready
func (h *BaseHandler) IsWorkDone(context handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	return false, ctrl.Result{}, nil
}

// PostWorkUpdateStatus does installation post-action
func (h BaseHandler) PostWorkUpdateStatus(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// PostWork does installation pre-action
func (h BaseHandler) PostWork(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// WorkCompletedUpdateStatus does the lifecycle completed Work status update
func (h *BaseHandler) WorkCompletedUpdateStatus(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// UpdateStatus does the lifecycle pre-Work status update
func (h BaseHandler) UpdateStatus(ctx handlerspi.HandlerContext, cond moduleapi.LifecycleCondition, state moduleapi.ModuleActionState) (ctrl.Result, error) {
	AppendCondition(h.ModuleCR, string(cond), cond)
	h.ModuleCR.Status.State = state
	if err := ctx.Client.Status().Update(context.TODO(), h.ModuleCR); err != nil {
		return util.NewRequeueWithShortDelay(), nil
	}
	return ctrl.Result{}, nil
}
