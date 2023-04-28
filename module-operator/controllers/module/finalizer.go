// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package module

import (
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/spi"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrl "sigs.k8s.io/controller-runtime"
)

const finalizerName = "platform.verrazzano.io/modulelifecycle.finalizer"

// GetName returns the name of the finalizer
func (r Reconciler) GetName() string {
	return finalizerName
}

// Cleanup garbage collects any related resources that were created by the controller
func (r Reconciler) Cleanup(spictx spi.ReconcileContext, u *unstructured.Unstructured) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

