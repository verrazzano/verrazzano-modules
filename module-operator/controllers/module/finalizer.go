// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package module

import (
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/result"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/spi/controllerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/statemachine"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const finalizerName = "module.platform.verrazzano.io/finalizer"

// GetName returns the name of the finalizer
func (r Reconciler) GetName() string {
	return finalizerName
}

// PreRemoveFinalizer is called when the resource is being deleted, before the finalizer
// is removed.  Use this method to delete Kubernetes resources, etc.
func (r Reconciler) PreRemoveFinalizer(spictx controllerspi.ReconcileContext, u *unstructured.Unstructured) result.Result {
	cr := &moduleapi.Module{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, cr); err != nil {
		return result.NewResult()
	}

	// Initialize the handler context
	handlerCtx, res := r.initHandlerCtx(spictx, cr)
	if res.ShouldRequeue() {
		return res
	}

	// Execute the state machine
	sm := statemachine.StateMachine{
		Handler: r.ModuleHandlerInfo.DeleteActionHandler,
		CR:      cr,
	}
	return funcExecuteStateMachine(handlerCtx, sm)
}

// PostRemoveFinalizer is called after the finalizer is successfully removed.
// This method does garbage collection and other tasks that can never return an error
func (r Reconciler) PostRemoveFinalizer(spictx controllerspi.ReconcileContext, u *unstructured.Unstructured) {
	// Delete the tracker used for this CR
	statemachine.DeleteTracker(u)
}
