// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

// Package statemachine provides a state machine for executing the flow of life-cycle operations. In this context,
// a life-cycle operation consists of the creation, mutation, and deletion of one or more Kubernetes resources. However,
// any reconciliation that requires work steps can use this state machine.
//
// The state machine calls Module handlers to do the work for each state.  The handlers must be idempotent
// since the state machine can be reset to the starting state at any time.  For example, if the CR generation changes
// while the state machine is executing, or if the operator pod restarts.
//
// There are a few reasons to use a state machine:
//   1. to optimize reconciliation so that the same code is not being executed repeatedly during the controller-runtime
//   Reconcile loop, which is called many times while waiting for Modules to finish work (e.g. Helm install).
//   2. to provide a structure for Module handler implementations, so all the code for the various Modules has a consistent
//   look and feel.
//   3. to make it easy for Module authors to implement the handlers for new Module, only overriding work steps that
//   are unique to that Module.

package statemachine

import (
	"fmt"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/result"
	handlerspi2 "github.com/verrazzano/verrazzano-modules/pkg/controller/spi/handlerspi"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
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
	CR      client.Object
	Handler handlerspi2.StateMachineHandler
}

// Execute runs the state machine starting at the state stored in the tracker.
// Each CR has a unique tracker for a given generation that tracks the current state.
// The state machine uses a work handler to implement module specific logic.  This
// state machine code is used by different types of controllers, such as the Module and
// ModuleAction controllers. It is important to note that if the CR generation increments,
// then a new tracker is created and the state machine starts from the beginning.
//
// During state machine execution, a result may be returned to indicate that the
// controller-runtime should requeue after a delay.  This is done when a handler is
// waiting for a resource or some other condition.
func (s *StateMachine) Execute(handlerContext handlerspi2.HandlerContext) result.Result {
	tracker := ensureTracker(s.CR, stateInit)

	workerName := s.Handler.GetWorkName()
	nsn := fmt.Sprintf("%s/%s", s.CR.GetNamespace(), s.CR.GetName())

	for tracker.state != stateEnd {
		switch tracker.state {
		case stateInit:
			tracker.state = stateCheckWorkNeeded

		case stateCheckWorkNeeded:
			needed, res := s.Handler.IsWorkNeeded(handlerContext)
			if res.ShouldRequeue() {
				return res
			}
			if !needed {
				tracker.state = stateEnd
			} else {
				tracker.state = statePreWorkUpdateStatus
			}

		case statePreWorkUpdateStatus:
			res := s.Handler.PreWorkUpdateStatus(handlerContext)
			if res.ShouldRequeue() {
				return res
			}
			tracker.state = statePreWork

		case statePreWork:
			tracker.preInstallTime = time.Now()
			handlerContext.Log.Progressf("Doing pre-%s for %s", workerName, nsn)
			res := s.Handler.PreWork(handlerContext)
			if res.ShouldRequeue() {
				return res
			}
			tracker.state = stateWorkUpdateStatus

		case stateWorkUpdateStatus:
			res := s.Handler.DoWorkUpdateStatus(handlerContext)
			if res.ShouldRequeue() {
				return res
			}
			tracker.state = stateWork

		case stateWork:
			handlerContext.Log.Progressf("Doing %s for %s", workerName, nsn)
			res := s.Handler.DoWork(handlerContext)
			if res.ShouldRequeue() {
				return res
			}
			tracker.state = stateWorkWaitDone

		case stateWorkWaitDone:
			done, res := s.Handler.IsWorkDone(handlerContext)
			if res.ShouldRequeue() {
				return res
			}
			if !done {
				return result.NewResultShortRequeueDelay()
			}
			tracker.state = statePostWorkUpdateStatus

		case statePostWorkUpdateStatus:
			res := s.Handler.PostWorkUpdateStatus(handlerContext)
			if res.ShouldRequeue() {
				return res
			}
			tracker.state = statePostWork

		case statePostWork:
			handlerContext.Log.Progressf("Doing post-%s for %s", workerName, nsn)
			res := s.Handler.PostWork(handlerContext)
			if res.ShouldRequeue() {
				return res
			}
			tracker.state = stateCompleteUpdateStatus

		case stateCompleteUpdateStatus:
			res := s.Handler.WorkCompletedUpdateStatus(handlerContext)
			if res.ShouldRequeue() {
				return res
			}
			tracker.state = stateEnd
		}
	}
	return result.NewResult()
}

// GetPreInstallTime returns the time right before preinstall is called
func GetPreInstallTime(cr client.Object) *time.Time {
	tracker := ensureTracker(cr, stateInit)
	return &tracker.preInstallTime
}
