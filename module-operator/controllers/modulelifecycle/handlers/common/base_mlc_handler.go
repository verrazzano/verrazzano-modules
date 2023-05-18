// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package common

import (
	"context"
	"github.com/verrazzano/verrazzano-modules/common/handlerspi"
	"github.com/verrazzano/verrazzano-modules/common/pkg/constants"
	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ handlerspi.StateMachineHandler = &BaseHandler{}

type BaseHandler struct {
	// Config is the handler configuration
	Config handlerspi.StateMachineHandlerConfig

	// HelmInfo has the helm information
	handlerspi.HelmInfo

	// ModuleCR is the ModuleLifecycle CR being handled
	ModuleCR *moduleapi.ModuleLifecycle

	// ChartDir is the helm chart directory (TODO remove this and use HelmInfo path)
	ChartDir string

	// ImagePullSecretKeyname is the Helm Value Key for the image pull secret for a chart
	ImagePullSecretKeyname string
}

func (h *BaseHandler) GetActionName() string {
	//TODO implement me
	return "unknown"
}

// Init initializes the handler with Helm chart information
func (h *BaseHandler) Init(_ handlerspi.HandlerContext, config handlerspi.StateMachineHandlerConfig) (ctrl.Result, error) {
	h.Config = config
	h.HelmInfo = config.HelmInfo
	h.ImagePullSecretKeyname = constants.GlobalImagePullSecName
	h.ModuleCR = config.CR.(*moduleapi.ModuleLifecycle)
	return ctrl.Result{}, nil
}

// IsActionNeeded returns true if install is needed
func (h BaseHandler) IsActionNeeded(ctx handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// PreActionUpdateStatus does the lifecycle pre-Action status update
func (h *BaseHandler) PreActionUpdateStatus(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// PreAction does the pre-action
func (h BaseHandler) PreAction(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// IsPreActionDone returns true if pre-action done
func (h BaseHandler) IsPreActionDone(ctx handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// ActionUpdateStatus does the lifecycle Action status update
func (h *BaseHandler) ActionUpdateStatus(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// DoAction installs the module using Helm
func (h *BaseHandler) DoAction(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// IsActionDone Indicates whether a module is installed and ready
func (h *BaseHandler) IsActionDone(context handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	return false, ctrl.Result{}, nil
}

// PostActionUpdateStatus does installation post-action
func (h BaseHandler) PostActionUpdateStatus(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// PostAction does installation pre-action
func (h BaseHandler) PostAction(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// IsPostActionDone returns true if post-action done
func (h BaseHandler) IsPostActionDone(ctx handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// CompletedActionUpdateStatus does the lifecycle completed Action status update
func (h *BaseHandler) CompletedActionUpdateStatus(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// UpdateStatus does the lifecycle pre-Action status update
func (h BaseHandler) UpdateStatus(ctx handlerspi.HandlerContext, cond moduleapi.LifecycleCondition, state moduleapi.ModuleLifecycleState) (ctrl.Result, error) {
	AppendCondition(h.ModuleCR, string(cond), cond)
	h.ModuleCR.Status.State = state
	if err := ctx.Client.Status().Update(context.TODO(), h.ModuleCR); err != nil {
		return util.NewRequeueWithShortDelay(), nil
	}
	return ctrl.Result{}, nil
}
