// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package result

import (
	"k8s.io/apimachinery/pkg/util/rand"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

type Result interface {
	ShouldRequeue() bool
	GetControllerResult() ctrl.Result
	GetError() error
	IsError() bool
}

type controllerResult struct {
	result ctrl.Result
	err    error
}

var _ Result = controllerResult{}

func NewRequeueWithShortDelay() Result {
	return NewRequeueWithDelay(1, 2, time.Second)
}

func NewResultRequeueIfError(err error) Result {
	if err != nil {
		return NewRequeueWithShortDelay()
	}
	return controllerResult{}
}

func NewResult() Result {
	return controllerResult{}
}

// NewRequeueWithDelay returns a new Result that will cause requeue after a short delay
func NewRequeueWithDelay(min int, max int, units time.Duration) Result {
	var random = rand.IntnRange(min, max)
	delayNanos := time.Duration(random) * units
	return controllerResult{result: ctrl.Result{Requeue: true, RequeueAfter: delayNanos}}
}

func (r controllerResult) ShouldRequeue() bool {
	return r.result.Requeue
}

func (r controllerResult) GetControllerResult() ctrl.Result {
	return r.result
}

func (r controllerResult) GetError() error {
	return r.err
}

func (r controllerResult) IsError() bool {
	return r.err.Error() != ""
}
