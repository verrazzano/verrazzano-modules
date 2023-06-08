// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package result

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewRequeueWithDelay tests the NewRequeueWithDelay func for the following use case
// GIVEN a request to NewRequeueWithDelay
// WHEN a min, max, time units are provided
// THEN a requeue result is returned with a delay within the specified bounds
func TestNewRequeueWithDelay(t *testing.T) {
	asserts := assert.New(t)
	requeueWithDelay := NewRequeueWithDelay(3, 5, time.Second)
	t.Logf("Requeue result: %v", requeueWithDelay)
	asserts.True(requeueWithDelay.ShouldRequeue())
	asserts.GreaterOrEqual(requeueWithDelay.GetCtrlRuntimeResult().RequeueAfter.Seconds(), (time.Duration(3) * time.Second).Seconds())
	asserts.LessOrEqual(requeueWithDelay.GetCtrlRuntimeResult().RequeueAfter.Seconds(), (time.Duration(5) * time.Second).Seconds())

	requeueWithDelay = NewRequeueWithDelay(3, 5, time.Second)
	t.Logf("Requeue result: %v", requeueWithDelay)
	asserts.True(requeueWithDelay.ShouldRequeue())
	asserts.GreaterOrEqual(requeueWithDelay.GetCtrlRuntimeResult().RequeueAfter.Seconds(), (time.Duration(3) * time.Second).Seconds())
	asserts.LessOrEqual(requeueWithDelay.GetCtrlRuntimeResult().RequeueAfter.Seconds(), (time.Duration(5) * time.Second).Seconds())

	requeueWithDelay = NewRequeueWithDelay(3, 5, time.Minute)
	t.Logf("Requeue result: %v", requeueWithDelay)
	asserts.True(requeueWithDelay.ShouldRequeue())
	asserts.GreaterOrEqual(requeueWithDelay.GetCtrlRuntimeResult().RequeueAfter.Seconds(), (time.Duration(3) * time.Minute).Seconds())
	asserts.LessOrEqual(requeueWithDelay.GetCtrlRuntimeResult().RequeueAfter.Seconds(), (time.Duration(5) * time.Minute).Seconds())
}
