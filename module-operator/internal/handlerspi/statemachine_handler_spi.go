// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package handlerspi

import (
	"github.com/verrazzano/verrazzano-modules/pkg/controller/result"
)

// StateMachineHandler is the interface called by the state machine to do module related work
type StateMachineHandler interface {
	// GetWorkName returns the work name
	GetWorkName() string

	// IsWorkNeeded returns true if work is needed for the Module
	IsWorkNeeded(context HandlerContext) (bool, result.Result)

	// PreWorkUpdateStatus does the pre-work status update
	PreWorkUpdateStatus(context HandlerContext) result.Result

	// PreWork does pre-work
	PreWork(context HandlerContext) result.Result

	// DoWorkUpdateStatus does the work status update
	DoWorkUpdateStatus(context HandlerContext) result.Result

	// DoWork does the work
	DoWork(context HandlerContext) result.Result

	// IsWorkDone returns true if work is done
	IsWorkDone(context HandlerContext) (bool, result.Result)

	// PostWorkUpdateStatus does the post-work status update
	PostWorkUpdateStatus(context HandlerContext) result.Result

	// PostWork does  post-work
	PostWork(context HandlerContext) result.Result

	// WorkCompletedUpdateStatus does the completed work status update
	WorkCompletedUpdateStatus(context HandlerContext) result.Result
}
