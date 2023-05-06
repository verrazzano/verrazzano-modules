// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package common

import (
	"context"
	actionspi "github.com/verrazzano/verrazzano-modules/common/actionspi"
	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano/platform-operator/constants"
	helmcomp "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/helm"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	ctrl "sigs.k8s.io/controller-runtime"
)

type BaseHandler struct {
	helmcomp.HelmComponent
	Config actionspi.HandlerConfig
	CR     *moduleapi.ModuleLifecycle
}

// Init initializes the component with Helm chart information
func (h *BaseHandler) Init(_ spi.ComponentContext, config actionspi.HandlerConfig) (ctrl.Result, error) {
	h.HelmComponent = helmcomp.HelmComponent{
		ReleaseName:             config.HelmInfo.HelmRelease.Name,
		ChartNamespace:          config.HelmInfo.HelmRelease.Namespace,
		ChartDir:                config.ChartDir,
		IgnoreNamespaceOverride: true,
		ImagePullSecretKeyname:  constants.GlobalImagePullSecName,
	}
	h.CR = config.CR.(*moduleapi.ModuleLifecycle)
	h.Config = config
	return ctrl.Result{}, nil
}

// UpdateStatus does the lifecycle pre-Action status update
func (h BaseHandler) UpdateStatus(ctx spi.ComponentContext, cond moduleapi.LifecycleCondition, state moduleapi.ModuleLifecycleState) (ctrl.Result, error) {
	AppendCondition(h.CR, string(cond), cond)
	h.CR.Status.State = state
	if err := ctx.Client().Status().Update(context.TODO(), h.CR); err != nil {
		return util.NewRequeueWithShortDelay(), nil
	}
	return ctrl.Result{}, nil
}
