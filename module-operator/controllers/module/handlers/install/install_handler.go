// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package install

import (
	"context"
	actionspi "github.com/verrazzano/verrazzano-modules/common/actionspi"
	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/handlers/common"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"

	moduleplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
)

type Handler struct {
	BaseHandler common.BaseHandler
}

var (
	_ actionspi.LifecycleActionHandler = &Handler{}
)

func NewHandler() actionspi.LifecycleActionHandler {
	return &Handler{}
}

// GetActionName returns the action name
func (h Handler) GetActionName() string {
	return "install"
}

// Init initializes the handler
func (h *Handler) Init(ctx spi.ComponentContext, config actionspi.HandlerConfig) (ctrl.Result, error) {
	return h.BaseHandler.Init(ctx, config, moduleplatform.InstallAction)
}

// IsActionNeeded returns true if install is needed
func (h Handler) IsActionNeeded(ctx spi.ComponentContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil

	//installed, err := vzhelm.IsReleaseInstalled(h.ReleaseName, h.chartDir)
	//if err != nil {
	//	ctx.Log().ErrorfThrottled("Error checking if Helm release installed for %s/%s", h.chartDir, h.ReleaseName)
	//	return true, ctrl.Result{}, err
	//}
	//return !installed, ctrl.Result{}, err
}

// PreActionUpdateStatus does the lifecycle pre-Action status update
func (h Handler) PreActionUpdateStatus(ctx spi.ComponentContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateStatus(ctx, moduleplatform.CondPreInstall, moduleplatform.ModuleStateReconciling)
}

// PreAction does installation pre-action
func (h Handler) PreAction(ctx spi.ComponentContext) (ctrl.Result, error) {
	// Update the spev version if it is not set
	if len(h.BaseHandler.ModuleCR.Spec.Version) == 0 {
		// Update spec version to match chart, always requeue to get ModuleCR with version
		h.BaseHandler.ModuleCR.Spec.Version = h.BaseHandler.Config.ChartInfo.Version
		if err := ctx.Client().Update(context.TODO(), h.BaseHandler.ModuleCR); err != nil {
			return util.NewRequeueWithShortDelay(), nil
		}
		// ALways reconcile so that we get a new tracker with the latest ModuleCR
		return util.NewRequeueWithDelay(1, 2, time.Second), nil
	}

	return ctrl.Result{}, nil
}

// IsPreActionDone returns true if pre-action done
func (h Handler) IsPreActionDone(ctx spi.ComponentContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// ActionUpdateStatus does the lifecycle Action status update
func (h Handler) ActionUpdateStatus(ctx spi.ComponentContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateStatus(ctx, moduleplatform.CondInstallStarted, moduleplatform.ModuleStateReconciling)
}

// DoAction installs the component using Helm
func (h Handler) DoAction(ctx spi.ComponentContext) (ctrl.Result, error) {
	return h.BaseHandler.DoAction(ctx)
}

// IsActionDone Indicates whether a component is installed and ready
func (h Handler) IsActionDone(ctx spi.ComponentContext) (bool, ctrl.Result, error) {
	return h.BaseHandler.IsActionDone(ctx)
}

// PostActionUpdateStatue does installation post-action status update
func (h Handler) PostActionUpdateStatus(ctx spi.ComponentContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// PostAction does installation post-action
func (h Handler) PostAction(ctx spi.ComponentContext) (ctrl.Result, error) {
	return h.BaseHandler.PostAction(ctx)
}

// IsPostActionDone returns true if post-action done
func (h Handler) IsPostActionDone(ctx spi.ComponentContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// CompletedActionUpdateStatus does the lifecycle pre-Action status update
func (h Handler) CompletedActionUpdateStatus(ctx spi.ComponentContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateStatusWithVersion(ctx, moduleplatform.CondInstallComplete, moduleplatform.ModuleStateReady, h.BaseHandler.ModuleCR.Spec.Version)
}
