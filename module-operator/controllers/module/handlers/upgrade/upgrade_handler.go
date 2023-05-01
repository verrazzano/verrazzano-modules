// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package upgrade

import (
	compspi "github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/action_spi"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"

	moduleplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
)
import (
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/handlers/common"
)

type Handler struct {
	BaseHandler common.BaseHandler
}

var (
	_ compspi.LifecycleActionHandler = &Handler{}
)

func NewHandler() compspi.LifecycleActionHandler {
	return &Handler{}
}

// GetStatusConditions returns the CR status conditions for various lifecycle stages
func (h *Handler) GetStatusConditions() compspi.StatusConditions {
	return compspi.StatusConditions{
		NotNeeded: moduleplatform.CondAlreadyUpgraded,
		PreAction: moduleplatform.CondPreUpgrade,
		DoAction:  moduleplatform.CondUpgradeStarted,
		Completed: moduleplatform.CondUpgradeComplete,
	}
}

// GetActionName returns the action name
func (h Handler) GetActionName() string {
	return "upgrade"
}

// Init initializes the handler
func (h *Handler) Init(ctx spi.ComponentContext, config compspi.HandlerConfig) (ctrl.Result, error) {
	return h.BaseHandler.Init(ctx, config, moduleplatform.UpgradeAction)
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

// PreAction does installation pre-action
func (h Handler) PreAction(ctx spi.ComponentContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// IsPreActionDone returns true if pre-action done
func (h Handler) IsPreActionDone(ctx spi.ComponentContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// DoAction installs the component using Helm
func (h Handler) DoAction(ctx spi.ComponentContext) (ctrl.Result, error) {
	return h.BaseHandler.DoAction(ctx)
}

// IsActionDone Indicates whether a component is installed and ready
func (h Handler) IsActionDone(ctx spi.ComponentContext) (bool, ctrl.Result, error) {
	return h.BaseHandler.IsActionDone(ctx)
}

// PostAction does installation post-action
func (h Handler) PostAction(ctx spi.ComponentContext) (ctrl.Result, error) {
	return h.BaseHandler.PostAction(ctx)
}

// IsPostActionDone returns true if post-action done
func (h Handler) IsPostActionDone(ctx spi.ComponentContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}
