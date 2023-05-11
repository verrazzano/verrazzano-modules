// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package common

import (
	"context"
	actionspi "github.com/verrazzano/verrazzano-modules/common/actionspi"
	"github.com/verrazzano/verrazzano-modules/common/pkg/constants"
	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type BaseHandler struct {
	// Config is the handler configuration
	Config actionspi.HandlerConfig

	// HelmInfo has the helm information
	actionspi.HelmInfo

	// ModuleCR is the ModuleLifecycle CR being handled
	ModuleCR *moduleapi.ModuleLifecycle

	// ChartDir is the helm chart directory (TODO remove this and use HelmInfo path)
	ChartDir string

	// ImagePullSecretKeyname is the Helm Value Key for the image pull secret for a chart
	ImagePullSecretKeyname string
}

// Init initializes the handler with Helm chart information
func (h *BaseHandler) Init(_ actionspi.HandlerContext, config actionspi.HandlerConfig) (ctrl.Result, error) {
	h.Config = config
	h.HelmInfo = config.HelmInfo
	h.ImagePullSecretKeyname = constants.GlobalImagePullSecName
	h.ModuleCR = config.CR.(*moduleapi.ModuleLifecycle)
	return ctrl.Result{}, nil
}

// UpdateStatus does the lifecycle pre-Action status update
func (h BaseHandler) UpdateStatus(ctx actionspi.HandlerContext, cond moduleapi.LifecycleCondition, state moduleapi.ModuleLifecycleState) (ctrl.Result, error) {
	AppendCondition(h.ModuleCR, string(cond), cond)
	h.ModuleCR.Status.State = state
	if err := ctx.Client.Status().Update(context.TODO(), h.ModuleCR); err != nil {
		return util.NewRequeueWithShortDelay(), nil
	}
	return ctrl.Result{}, nil
}
