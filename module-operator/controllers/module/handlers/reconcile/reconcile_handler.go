// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package reconcile

import (
	"context"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/handlers/common"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/handlerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"

	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
)

type Handler struct {
	common.BaseHandler
}

var (
	_ handlerspi.StateMachineHandler = &Handler{}
)

func NewHandler() handlerspi.StateMachineHandler {
	return &Handler{}
}

// GetWorkName returns the work name
func (h Handler) GetWorkName() string {
	return string(moduleapi.ReconcileAction)
}

// Init initializes the handler
func (h *Handler) Init(ctx handlerspi.HandlerContext, config handlerspi.StateMachineHandlerConfig) (ctrl.Result, error) {
	return h.BaseHandler.Init(ctx, config, moduleapi.ReconcileAction)
}

// PreWorkUpdateStatus does the lifecycle pre-work status update
func (h Handler) PreWorkUpdateStatus(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateStatus(ctx, moduleapi.CondReconciling, moduleapi.ModuleStateReconciling)
}

// PreWork does installation pre-work
func (h Handler) PreWork(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	// Update the spec version if it is not set
	if len(h.BaseHandler.ModuleCR.Spec.Version) == 0 {
		// Update spec version to match chart, always requeue to get ModuleCR with version
		h.BaseHandler.ModuleCR.Spec.Version = h.BaseHandler.Config.ChartInfo.Version
		if err := ctx.Client.Update(context.TODO(), h.BaseHandler.ModuleCR); err != nil {
			return util.NewRequeueWithShortDelay(), nil
		}
		// ALways reconcile so that we get a new tracker with the latest ModuleCR
		return util.NewRequeueWithDelay(1, 2, time.Second), nil
	}

	return ctrl.Result{}, nil
}

// WorkCompletedUpdateStatus does the lifecycle pre-Action status update
func (h Handler) WorkCompletedUpdateStatus(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return h.BaseHandler.UpdateDoneStatus(ctx, moduleapi.CondReconcilingComplete, moduleapi.ModuleStateReady, h.BaseHandler.ModuleCR.Spec.Version)
}
