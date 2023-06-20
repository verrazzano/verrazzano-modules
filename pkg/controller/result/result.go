// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package result

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

type Result interface {
	ShouldRequeue() bool
	GetCtrlRuntimeResult() ctrl.Result
	GetError() error
	IsError() bool
}

type controllerResult struct {
	result ctrl.Result
	err    error
}

var _ Result = controllerResult{}

func NewResult() Result {
	return controllerResult{}
}

// NewResultShortRequeueDelay returns a new Result that will cause requeue after a short delay
func NewResultShortRequeueDelay() Result {
	return NewBuilder().ShortDelay().Build()
}

// NewResultRequeueDelay returns a new Result that will cause requeue after the specified delay
func NewResultRequeueDelay(min int, max int, units time.Duration) Result {
	return NewBuilder().Delay(min, max, units).Build()
}

// NewResultShortRequeueDelayIfError returns a new Result that will cause requeue after a short delay if there is an error
func NewResultShortRequeueDelayIfError(err error) Result {
	b := NewBuilder()
	if err != nil {
		b.Error(err).ShortDelay()
	}
	return b.Build()
}

// NewResultShortRequeueDelayWithError returns a new Result that will cause requeue after a short delay
func NewResultShortRequeueDelayWithError(err error) Result {
	b := NewBuilder()
	b.Error(err).ShortDelay()
	return b.Build()
}

func (r controllerResult) ShouldRequeue() bool {
	return r.result.Requeue
}

func (r controllerResult) GetCtrlRuntimeResult() ctrl.Result {
	return r.result
}

func (r controllerResult) GetError() error {
	return r.err
}

func (r controllerResult) IsError() bool {
	return r.err != nil
}
