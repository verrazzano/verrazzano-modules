// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

// Package statemachine provides a state machine for executing the flow of life-cycle actions.
// The current state of the state machine is stored in a tracker that is unique for each ModuleCR generation.
// This allows the state machine to be initialized and called several times until all states have executed,
// during a controller-runtime reconcile loop, where the Reconcile method is called repeatedly.
package statemachine

import (
	"fmt"
	"github.com/verrazzano/verrazzano-modules/common/actionspi"
	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// state identifies the state of a component during action
type state string

const (
	// stateInit is the state when a component is initialized
	stateInit state = "stateInit"

	// stateCheckActionNeeded is the state to check if action is needed
	stateCheckActionNeeded state = "stateCheckActionNeeded"

	// statePreActionUpdateStatus is the state when the status is updated to start pre action
	statePreActionUpdateStatus state = "statePreActionUpdateStatus"

	// statePreAction is the state when a component does a pre-action
	statePreAction state = "statePreAction"

	// statePreActionWaitDone is the state when a component is waiting for pre-action to be done
	statePreActionWaitDone state = "statePreActionWaitDone"

	// stateActionUpdateStatus is the state when the status is updated to start action
	stateActionUpdateStatus state = "stateActionUpdateStatus"

	// stateAction is the state where a component does an action
	stateAction state = "stateAction"

	// stateActionWaitDone is the state when a component is waiting for action to be done
	stateActionWaitDone state = "stateActionWaitDone"

	// statePostActionUpdateStatus is the state when the status is updated to start post-action
	statePostActionUpdateStatus state = "statePostActionUpdateStatus"

	// statePostAction is the state when a component does a post-action
	statePostAction state = "statePostAction"

	// statePostActionWaitDone is the state when a component is waiting for post-action to be done
	statePostActionWaitDone state = "statePostActionWaitDone"

	// stateCompleteUpdateStatus is the state when the status is updated to completed
	stateCompleteUpdateStatus state = "stateCompleteUpdateStatus"

	// stateEnd is the terminal state
	stateEnd state = "stateEnd"
)

// StateMachine contains the fields needed to execute the state machine.
type StateMachine struct {
	*runtime.Scheme
	CR       client.Object
	HelmInfo *actionspi.HelmInfo
	Handler  actionspi.LifecycleActionHandler
}

// Execute runs the state machine starting at the state stored in the tracker.
// Each CR has a unique tracker for a given generation that tracks the current state.
// The state machine uses an action handler to implement module specific logic.  This
// state machine code is used by different types of controllers, such as the Module and
// ModuleLifeCycle controllers.
//
// During state machine execution, a result may be returned to indicate that the
// controller-runtime should requeue after a delay.  This is done when a handler is
// waiting for a resource or some other condition.
//
// It is important to note that if the CR generation increments, then a new tracker is created
// and the state machine starts from the beginning.
func (s *StateMachine) Execute(handlerContext actionspi.HandlerContext) ctrl.Result {
	tracker := ensureTracker(s.CR, stateInit)

	actionName := s.Handler.GetActionName()
	nsn := fmt.Sprintf("%s/%s", s.CR.GetNamespace(), s.CR.GetName())

	for tracker.state != stateEnd {
		switch tracker.state {
		case stateInit:
			// Init the Handler
			config := actionspi.HandlerConfig{
				HelmInfo: *s.HelmInfo,
				CR:       s.CR,
				Scheme:   s.Scheme,
			}
			res, err := s.Handler.Init(handlerContext, config)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			tracker.state = stateCheckActionNeeded

		case stateCheckActionNeeded:
			needed, res, err := s.Handler.IsActionNeeded(handlerContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			if !needed {
				tracker.state = stateEnd
			} else {
				tracker.state = statePreActionUpdateStatus
			}

		case statePreActionUpdateStatus:
			res, err := s.Handler.PreActionUpdateStatus(handlerContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			tracker.state = statePreAction

		case statePreAction:
			handlerContext.Log.Progressf("Doing pre-%s for %s", actionName, nsn)
			res, err := s.Handler.PreAction(handlerContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			tracker.state = statePreActionWaitDone

		case statePreActionWaitDone:
			done, res, err := s.Handler.IsPreActionDone(handlerContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			if !done {
				return util.NewRequeueWithShortDelay()
			}
			tracker.state = stateActionUpdateStatus

		case stateActionUpdateStatus:
			res, err := s.Handler.ActionUpdateStatus(handlerContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			tracker.state = stateAction

		case stateAction:
			handlerContext.Log.Progressf("Doing %s for %s", actionName, nsn)
			res, err := s.Handler.DoAction(handlerContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			tracker.state = stateActionWaitDone

		case stateActionWaitDone:
			done, res, err := s.Handler.IsActionDone(handlerContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			if !done {
				return util.NewRequeueWithShortDelay()
			}
			tracker.state = statePostActionUpdateStatus

		case statePostActionUpdateStatus:
			res, err := s.Handler.PostActionUpdateStatus(handlerContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			tracker.state = statePostAction

		case statePostAction:
			handlerContext.Log.Progressf("Doing post-%s for %s", actionName, nsn)
			res, err := s.Handler.PostAction(handlerContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			tracker.state = statePostActionWaitDone

		case statePostActionWaitDone:
			done, res, err := s.Handler.IsPostActionDone(handlerContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			if !done {
				return util.NewRequeueWithShortDelay()
			}
			tracker.state = stateCompleteUpdateStatus

		case stateCompleteUpdateStatus:
			res, err := s.Handler.CompletedActionUpdateStatus(handlerContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			tracker.state = stateEnd
		}
	}
	return ctrl.Result{}
}
