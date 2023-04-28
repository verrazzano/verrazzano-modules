// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package module

import (
	compspi "github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/action_spi"
	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	"github.com/verrazzano/verrazzano-modules/common/pkg/k8s"
	moduleplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	vzspi "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	ctrl "sigs.k8s.io/controller-runtime"
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

	// stateStartPreActionUpdate is the state when the status is updated to start pre action
	stateStartPreActionUpdate state = "stateStartPreActionUpdate"

	// stateStartActionUpdate is the state when the status is updated to start action
	stateStartActionUpdate state = "stateStartActionUpdate"

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

	// stateCompleteUpdate is the state when the status is updated to completed
	stateCompleteUpdate state = "stateCompleteUpdate"

	// stateEnd is the terminal state
	stateEnd state = "stateEnd"
)

type stateMachineContext struct {
	cr        *moduleplatform.Module
	tracker   *stateTracker
	chartInfo *compspi.HelmInfo
	action    compspi.LifecycleActionHandler
}

func (r *Reconciler) doStateMachine(spiCtx vzspi.ComponentContext, s stateMachineContext) ctrl.Result {
	actionName := "action-placeholder"
	compContext := spiCtx.Init("component").Operation(string(actionName))
	nsn := k8s.GetNamespacedNameString(s.cr.ObjectMeta)

	for s.tracker.state != stateEnd {
		switch s.tracker.state {
		case stateInit:
			res, err := s.action.Init(compContext, s.chartInfo, s.cr.Namespace, s.cr)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			s.tracker.state = stateCheckActionNeeded

		case stateCheckActionNeeded:
			needed, res, err := s.action.IsActionNeeded(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			if needed {
				s.tracker.state = stateStartPreActionUpdate
			} else {
				s.tracker.state = stateActionNotNeededUpdate
			}

		case stateActionNotNeededUpdate:
			//cond := s.action.GetStatusConditions().NotNeeded
			//if err := UpdateStatus(r.Client, s.cr, string(cond), cond); err != nil {
			//	return util.NewRequeueWithShortDelay()
			//}
			s.tracker.state = stateEnd

		case stateStartPreActionUpdate:
			//cond := s.action.GetStatusConditions().PreAction
			//if err := UpdateStatus(r.Client, s.cr, string(cond), cond); err != nil {
			//	return util.NewRequeueWithShortDelay()
			//}
			s.tracker.state = statePreAction

		case statePreAction:
			spiCtx.Log().Progressf("Doing pre-%s for %s", actionName, nsn)
			res, err := s.action.PreAction(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			s.tracker.state = statePreActionWaitDone

		case statePreActionWaitDone:
			done, res, err := s.action.IsPreActionDone(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			if !done {
				return util.NewRequeueWithShortDelay()
			}
			s.tracker.state = stateStartActionUpdate

		case stateStartActionUpdate:
			//cond := s.action.GetStatusConditions().DoAction
			//if err := UpdateStatus(r.Client, s.cr, string(cond), cond); err != nil {
			//	return util.NewRequeueWithShortDelay()
			//}
			s.tracker.state = stateAction

		case stateAction:
			spiCtx.Log().Progressf("Doing %s for %s", actionName, nsn)
			res, err := s.action.DoAction(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			s.tracker.state = stateActionWaitDone

		case stateActionWaitDone:
			done, res, err := s.action.IsActionDone(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			if !done {
				return util.NewRequeueWithShortDelay()
			}
			s.tracker.state = statePostAction

		case statePostAction:
			spiCtx.Log().Progressf("Doing post-%s for %s", actionName, nsn)
			res, err := s.action.PostAction(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			s.tracker.state = statePostActionWaitDone

		case statePostActionWaitDone:
			done, res, err := s.action.IsPostActionDone(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			if !done {
				return util.NewRequeueWithShortDelay()
			}
			s.tracker.state = stateCompleteUpdate

		case stateCompleteUpdate:
			//cond := s.action.GetStatusConditions().Completed
			//if err := UpdateStatus(r.Client, s.cr, string(cond), cond); err != nil {
			//	return util.NewRequeueWithShortDelay()
			//}
			spiCtx.Log().Progressf("Successfully completed %s for %s", actionName, nsn)

			s.tracker.state = stateEnd
		}
	}
	return ctrl.Result{}
}
