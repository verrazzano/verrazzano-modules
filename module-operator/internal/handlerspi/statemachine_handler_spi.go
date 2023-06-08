// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package handlerspi

import (
	ctrl "sigs.k8s.io/controller-runtime"
)

// StateMachineHandler is the interface called by the state machine to do module related work
type StateMachineHandler interface {
	// GetWorkName returns the work name
	GetWorkName() string

	// IsWorkNeeded returns true if work is needed for the Module
	IsWorkNeeded(context HandlerContext) (bool, ctrl.Result, error)

	// PreWorkUpdateStatus does the pre-work status update
	PreWorkUpdateStatus(context HandlerContext) (ctrl.Result, error)

	// PreWork does pre-work
	PreWork(context HandlerContext) (ctrl.Result, error)

	// DoWorkUpdateStatus does the work status update
	DoWorkUpdateStatus(context HandlerContext) (ctrl.Result, error)

	// DoWork does the work
	DoWork(context HandlerContext) (ctrl.Result, error)

	// IsWorkDone returns true if work is done
	IsWorkDone(context HandlerContext) (bool, ctrl.Result, error)

	// PostWorkUpdateStatus does the post-work status update
	PostWorkUpdateStatus(context HandlerContext) (ctrl.Result, error)

	// PostWork does  post-work
	PostWork(context HandlerContext) (ctrl.Result, error)

	// WorkCompletedUpdateStatus does the completed work status update
	WorkCompletedUpdateStatus(context HandlerContext) (ctrl.Result, error)
}
