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
}

type ControllerResult struct {
	result ctrl.Result
	err    error
}

var _ Result = ControllerResult{}

func NewRequeueWithShortDelay() Result {
	return NewRequeueWithDelay(1, 2, time.Second)
}

// NewRequeueWithDelay returns a new Result that will cause requeue after a short delay
func NewRequeueWithDelay(min int, max int, units time.Duration) ControllerResult {
	var random = rand.IntnRange(min, max)
	delayNanos := time.Duration(random) * units
	return ControllerResult{result: ctrl.Result{Requeue: true, RequeueAfter: delayNanos}}
}

func (r ControllerResult) ShouldRequeue() bool {
	return r.result.Requeue
}

func (r ControllerResult) GetControllerResult() ctrl.Result {
	return r.result
}

func (r ControllerResult) GetError() error {
	return r.err
}
