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

type BaseHandler struct {
	// Config is the handler configuration
	Config handlerspi.StateMachineHandlerConfig

	// HelmInfo has the helm information
	handlerspi.HelmInfo

	// ModuleCR is the Module CR being handled
	ModuleCR *moduleapi.Module

	// ChartDir is the helm chart directory (TODO remove this and use HelmInfo path)
	ChartDir string

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

// UpdateDoneStatus does the lifecycle status update with the action is done
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
