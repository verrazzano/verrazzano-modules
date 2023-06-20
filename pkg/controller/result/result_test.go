// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package result

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// NewResult tests the NewResult method
// GIVEN a request to NewResult
// WHEN NewResult is called
// THEN a requeue result is returned with correct settings
func TestNewResult(t *testing.T) {
	asserts := assert.New(t)
	r := NewResult()
	asserts.False(r.ShouldRequeue())
	asserts.Zero(r.GetCtrlRuntimeResult().RequeueAfter.Seconds())
}

// NewResultShortRequeueDelay tests the NewResultShortRequeueDelay method
// GIVEN a request to NewResultShortRequeueDelay
// WHEN a min, max, time units are provided
// THEN a requeue result is returned with a delay within the specified bounds
func TestNewResultShortRequeueDelay(t *testing.T) {
	asserts := assert.New(t)
	requeueWithDelay := NewResultShortRequeueDelay()
	asserts.True(requeueWithDelay.ShouldRequeue())
	asserts.GreaterOrEqual(requeueWithDelay.GetCtrlRuntimeResult().RequeueAfter.Seconds(), (time.Duration(1) * time.Second).Seconds())
	asserts.LessOrEqual(requeueWithDelay.GetCtrlRuntimeResult().RequeueAfter.Seconds(), (time.Duration(2) * time.Second).Seconds())
}

// TestNewResultShortRequeueDelayWithError tests the NewResultShortRequeueDelayWithError method
// GIVEN a request to TestNewResultShortRequeueDelayWithError
// WHEN TestNewResultShortRequeueDelayWithError with and without and error
// THEN should requeue returns the correct result
func TestNewResultShortRequeueDelayWithError(t *testing.T) {
	asserts := assert.New(t)
	r := NewResultShortRequeueDelayWithError(nil)
	asserts.True(r.ShouldRequeue())
	asserts.False(r.IsError())
	asserts.NoError(r.GetError())

	r = NewResultShortRequeueDelayWithError(errors.New("err"))
	asserts.True(r.ShouldRequeue())
	asserts.True(r.IsError())
	asserts.Error(r.GetError())
}

// TestNewResultShortRequeueDelayIfError tests the NewResultShortRequeueDelayIfError method
// GIVEN a request to TestNewResultShortRequeueDelayIfError
// WHEN TestNewResultShortRequeueDelayIfError with and without and error
// THEN should requeue returns the correct result
func TestNewResultShortRequeueDelayIfError(t *testing.T) {
	asserts := assert.New(t)
	r := NewResultShortRequeueDelayIfError(nil)
	asserts.False(r.ShouldRequeue())
	asserts.False(r.IsError())
	asserts.NoError(r.GetError())

	r = NewResultShortRequeueDelayIfError(errors.New("err"))
	asserts.True(r.ShouldRequeue())
	asserts.True(r.IsError())
	asserts.Error(r.GetError())
}

// TestNewRequeueWithDelay tests the NewResultRequeueDelay func for the following use case
// GIVEN a request to NewResultRequeueDelay
// WHEN a min, max, time units are provided
// THEN a requeue result is returned with a delay within the specified bounds
func TestNewRequeueWithDelay(t *testing.T) {
	asserts := assert.New(t)
	requeueWithDelay := NewResultRequeueDelay(3, 5, time.Millisecond)
	asserts.True(requeueWithDelay.ShouldRequeue())
	asserts.GreaterOrEqual(requeueWithDelay.GetCtrlRuntimeResult().RequeueAfter.Milliseconds(), (time.Duration(3) * time.Millisecond).Milliseconds())
	asserts.LessOrEqual(requeueWithDelay.GetCtrlRuntimeResult().RequeueAfter.Milliseconds(), (time.Duration(5) * time.Millisecond).Milliseconds())

	requeueWithDelay = NewResultRequeueDelay(3, 5, time.Second)
	asserts.True(requeueWithDelay.ShouldRequeue())
	asserts.GreaterOrEqual(requeueWithDelay.GetCtrlRuntimeResult().RequeueAfter.Seconds(), (time.Duration(3) * time.Second).Seconds())
	asserts.LessOrEqual(requeueWithDelay.GetCtrlRuntimeResult().RequeueAfter.Seconds(), (time.Duration(5) * time.Second).Seconds())

	requeueWithDelay = NewResultRequeueDelay(3, 5, time.Minute)
	asserts.True(requeueWithDelay.ShouldRequeue())
	asserts.GreaterOrEqual(requeueWithDelay.GetCtrlRuntimeResult().RequeueAfter.Seconds(), (time.Duration(3) * time.Minute).Seconds())
	asserts.LessOrEqual(requeueWithDelay.GetCtrlRuntimeResult().RequeueAfter.Seconds(), (time.Duration(5) * time.Minute).Seconds())
}
