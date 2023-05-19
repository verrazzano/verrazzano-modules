// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

// Package statemachine provides a state machine for executing the flow of life-cycle works.
// The current state of the state machine is stored in a tracker that is unique for each ModuleCR generation.
// This allows the state machine to be initialized and called several times until all states have executed,
// during a controller-runtime reconcile loop, where the Reconcile method is called repeatedly.
package statemachine

import (
	"fmt"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/handlerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/util"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// state identifies the state of a component during work
type state string

const (
	// stateInit is the state when a component is initialized
	stateInit state = "stateInit"

	// stateCheckWorkNeeded is the state to check if work is needed
	stateCheckWorkNeeded state = "stateCheckWorkNeeded"

	// statePreWorkUpdateStatus is the state when the status is updated to start pre work
	statePreWorkUpdateStatus state = "statePreWorkUpdateStatus"

	// statePreWork is the state when a component does a pre-work
	statePreWork state = "statePreWork"

	// stateWorkUpdateStatus is the state when the status is updated to start work
	stateWorkUpdateStatus state = "stateWorkUpdateStatus"

	// stateWork is the state where a component does an work
	stateWork state = "stateWork"

	// stateWorkWaitDone is the state when a component is waiting for work to be done
	stateWorkWaitDone state = "stateWorkWaitDone"

	// statePostWorkUpdateStatus is the state when the status is updated to start post-work
	statePostWorkUpdateStatus state = "statePostWorkUpdateStatus"

	// statePostWork is the state when a component does a post-work
	statePostWork state = "statePostWork"

	// stateCompleteUpdateStatus is the state when the status is updated to completed
	stateCompleteUpdateStatus state = "stateCompleteUpdateStatus"

	// stateEnd is the terminal state
	stateEnd state = "stateEnd"
)

// StateMachine contains the fields needed to execute the state machine.
type StateMachine struct {
	*runtime.Scheme
	CR       client.Object
	HelmInfo *handlerspi.HelmInfo
	Handler  handlerspi.StateMachineHandler
}

// Execute runs the state machine starting at the state stored in the tracker.
// Each CR has a unique tracker for a given generation that tracks the current state.
// The state machine uses a work handler to implement module specific logic.  This
// state machine code is used by different types of controllers, such as the Module and
// ModuleLifeCycle controllers.
//
// During state machine execution, a result may be returned to indicate that the
// controller-runtime should requeue after a delay.  This is done when a handler is
// waiting for a resource or some other condition.
//
// It is important to note that if the CR generation increments, then a new tracker is created
// and the state machine starts from the beginning.
func (s *StateMachine) Execute(handlerContext handlerspi.HandlerContext) ctrl.Result {
	tracker := ensureTracker(s.CR, stateInit)

	workerName := s.Handler.GetWorkName()
	nsn := fmt.Sprintf("%s/%s", s.CR.GetNamespace(), s.CR.GetName())

	for tracker.state != stateEnd {
		switch tracker.state {
		case stateInit:
			// Init the Handler
			config := handlerspi.StateMachineHandlerConfig{
				HelmInfo: *s.HelmInfo,
				CR:       s.CR,
				Scheme:   s.Scheme,
			}
			res, err := s.Handler.Init(handlerContext, config)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			tracker.state = stateCheckWorkNeeded

		case stateCheckWorkNeeded:
			needed, res, err := s.Handler.IsWorkNeeded(handlerContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			if !needed {
				tracker.state = stateEnd
			} else {
				tracker.state = statePreWorkUpdateStatus
			}

		case statePreWorkUpdateStatus:
			res, err := s.Handler.PreWorkUpdateStatus(handlerContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			tracker.state = statePreWork

		case statePreWork:
			handlerContext.Log.Progressf("Doing pre-%s for %s", workerName, nsn)
			res, err := s.Handler.PreWork(handlerContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			tracker.state = stateWorkUpdateStatus

		case stateWorkUpdateStatus:
			res, err := s.Handler.DoWorkUpdateStatus(handlerContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			tracker.state = stateWork

		case stateWork:
			handlerContext.Log.Progressf("Doing %s for %s", workerName, nsn)
			res, err := s.Handler.DoWork(handlerContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			tracker.state = stateWorkWaitDone

		case stateWorkWaitDone:
			done, res, err := s.Handler.IsWorkDone(handlerContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			if !done {
				return util.NewRequeueWithShortDelay()
			}
			tracker.state = statePostWorkUpdateStatus

		case statePostWorkUpdateStatus:
			res, err := s.Handler.PostWorkUpdateStatus(handlerContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			tracker.state = statePostWork

		case statePostWork:
			handlerContext.Log.Progressf("Doing post-%s for %s", workerName, nsn)
			res, err := s.Handler.PostWork(handlerContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			tracker.state = stateCompleteUpdateStatus

		case stateCompleteUpdateStatus:
			res, err := s.Handler.WorkCompletedUpdateStatus(handlerContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			tracker.state = stateEnd
		}
	}
	return ctrl.Result{}
}
