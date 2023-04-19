// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package controllerutils

import (
	"k8s.io/apimachinery/pkg/util/rand"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

func NewRequeueWithShortDelay() ctrl.Result {
	return NewRequeueWithDelay(1, 2, time.Second)
}

// NewRequeueWithDelay returns a new Result that will cause requeue after a short delay
func NewRequeueWithDelay(min int, max int, units time.Duration) ctrl.Result {
	var seconds = rand.IntnRange(min, max)
	delaySecs := time.Duration(seconds) * units
	return ctrl.Result{Requeue: true, RequeueAfter: delaySecs}
}

// ShouldRequeue returns true if requeue is needed
func ShouldRequeue(r ctrl.Result) bool {
	return r.Requeue || r.RequeueAfter > 0
}

// DeriveResult will derive a new result from the input result and error
func DeriveResult(res ctrl.Result, err error) ctrl.Result {
	if ShouldRequeue(res) {
		// Always have at least a small delay to avoid pinning cpu
		if res.RequeueAfter == 0 {
			return NewRequeueWithShortDelay()
		}
		// Input result had a specific RequeueAfter, retain the setting
		return res
	}
	if err != nil {
		return NewRequeueWithShortDelay()
	}
	return ctrl.Result{}
}
