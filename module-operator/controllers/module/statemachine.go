// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package module

import (
	"context"
	compspi "github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/action_spi"
	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	"github.com/verrazzano/verrazzano-modules/common/pkg/k8s"
	moduleplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	vzspi "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

// state identifies the state of a component during action
type state string

const (
	// stateInit is the state when a component is initialized
	stateInit state = "stateInit"

	// stateCheckActionNeeded is the state to check if action is needed
	stateCheckActionNeeded state = "stateCheckActionNeeded"

	// stateActionNotNeededUpdate is the state when the status is updated to not needed
	stateActionNotNeededUpdate state = "stateActionNotNeededUpdate"

	// statePreActionUpdateStatus is the state when the status is updated to start pre action
	statePreActionUpdateStatus state = "statePreActionUpdateStatus"

	// stateActionUpdateStatus is the state when the status is updated to start action
	stateActionUpdateStatus state = "stateActionUpdateStatus"

	// statePreAction is the state when a component does a pre-action
	statePreAction state = "statePreAction"

	// statePreActionWaitDone is the state when a component is waiting for pre-action to be done
	statePreActionWaitDone state = "statePreActionWaitDone"

	// stateAction is the state where a component does an action
	stateAction state = "stateAction"

	// stateActionWaitDone is the state when a component is waiting for action to be done
	stateActionWaitDone state = "stateActionWaitDone"

	// statePostAction is the state when a component does a post-action
	statePostAction state = "statePostAction"

	// statePostActionWaitDone is the state when a component is waiting for post-action to be done
	statePostActionWaitDone state = "statePostActionWaitDone"

	// stateCompleteUpdateStatus is the state when the status is updated to completed
	stateCompleteUpdateStatus state = "stateCompleteUpdateStatus"

	// stateEnd is the terminal state
	stateEnd state = "stateEnd"
)

type stateMachineContext struct {
	cr       *moduleplatform.Module
	tracker  *stateTracker
	helmInfo *compspi.HelmInfo
	handler  compspi.LifecycleActionHandler
}

func (r *Reconciler) doStateMachine(spiCtx vzspi.ComponentContext, s stateMachineContext) ctrl.Result {
	actionName := s.handler.GetActionName()
	compContext := spiCtx.Init("component").Operation(actionName)
	nsn := k8s.GetNamespacedNameString(s.cr.ObjectMeta)

	for s.tracker.state != stateEnd {
		switch s.tracker.state {
		case stateInit:
			if len(s.cr.Spec.Version) == 0 {
				// Update spec version to match chart, always requeue to get CR with version
				s.cr.Spec.Version = s.helmInfo.ChartInfo.Version
				if err := r.Client.Update(context.TODO(), s.cr); err != nil {
					return util.NewRequeueWithShortDelay()
				}
				// ALways reconcile
				return util.NewRequeueWithDelay(1, 2, time.Second)
			}

			// Init the handler
			config := compspi.HandlerConfig{
				HelmInfo: *s.helmInfo,
				CR:       s.cr,
				Scheme:   r.Scheme,
			}
			res, err := s.handler.Init(compContext, config)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			s.tracker.state = stateCheckActionNeeded

		case stateCheckActionNeeded:
			needed, res, err := s.handler.IsActionNeeded(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			if !needed {
				s.tracker.state = stateActionNotNeededUpdate
			} else {
				s.tracker.state = statePreActionUpdateStatus
			}

		case stateActionNotNeededUpdate:
			s.cr.Status.State = moduleplatform.ModuleStateReady
			if err := r.Client.Status().Update(context.TODO(), s.cr); err != nil {
				return util.NewRequeueWithShortDelay()
			}
			s.tracker.state = stateEnd

		case statePreActionUpdateStatus:
			cond := s.handler.GetStatusConditions().PreAction
			AppendCondition(s.cr, string(cond), cond)
			s.cr.Status.State = moduleplatform.ModuleStateReconciling
			if err := r.Client.Status().Update(context.TODO(), s.cr); err != nil {
				return util.NewRequeueWithShortDelay()
			}
			s.tracker.state = statePreAction

		case statePreAction:
			spiCtx.Log().Progressf("Doing pre-%s for %s", actionName, nsn)
			res, err := s.handler.PreAction(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			s.tracker.state = statePreActionWaitDone

		case statePreActionWaitDone:
			done, res, err := s.handler.IsPreActionDone(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			if !done {
				return util.NewRequeueWithShortDelay()
			}
			s.tracker.state = stateActionUpdateStatus

		case stateActionUpdateStatus:
			cond := s.handler.GetStatusConditions().DoAction
			AppendCondition(s.cr, string(cond), cond)
			if err := r.Client.Status().Update(context.TODO(), s.cr); err != nil {
				return util.NewRequeueWithShortDelay()
			}
			s.tracker.state = stateAction

		case stateAction:
			spiCtx.Log().Progressf("Doing %s for %s", actionName, nsn)
			res, err := s.handler.DoAction(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			s.tracker.state = stateActionWaitDone

		case stateActionWaitDone:
			done, res, err := s.handler.IsActionDone(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			if !done {
				return util.NewRequeueWithShortDelay()
			}
			s.tracker.state = statePostAction

		case statePostAction:
			spiCtx.Log().Progressf("Doing post-%s for %s", actionName, nsn)
			res, err := s.handler.PostAction(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			s.tracker.state = statePostActionWaitDone

		case statePostActionWaitDone:
			done, res, err := s.handler.IsPostActionDone(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			if !done {
				return util.NewRequeueWithShortDelay()
			}
			s.tracker.state = stateCompleteUpdateStatus

		case stateCompleteUpdateStatus:
			cond := s.handler.GetStatusConditions().Completed
			AppendCondition(s.cr, string(cond), cond)
			s.cr.Status.State = moduleplatform.ModuleStateReady
			s.cr.Status.Version = s.cr.Spec.Version
			if err := r.Client.Status().Update(context.TODO(), s.cr); err != nil {
				return util.NewRequeueWithShortDelay()
			}
			spiCtx.Log().Progressf("Successfully completed %s for %s", actionName, nsn)

			s.tracker.state = stateEnd
		}
	}
	return ctrl.Result{}
}
