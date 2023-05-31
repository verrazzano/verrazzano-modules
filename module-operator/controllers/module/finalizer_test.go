// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package module

import (
	"context"
	"github.com/stretchr/testify/assert"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/handlerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/base/controllerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

type FinalizerReconciler struct {
	Scheme      *runtime.Scheme
	HandlerInfo handlerspi.ModuleHandlerInfo
	ModuleClass moduleapi.ModuleClassType
}

var _ controllerspi.Reconciler = &FinalizerReconciler{}

// TestFinalizer tests that the finalizer implementation works correctly
// GIVEN a Finalizer
// WHEN the Finalizer methods are called
// THEN ensure that they work correctly
func TestFinalizer(t *testing.T) {
	asserts := assert.New(t)

	r := Reconciler{}
	asserts.Equal(finalizerName, r.GetName())

	rctx := controllerspi.ReconcileContext{
		Log:       vzlog.DefaultLogger(),
		ClientCtx: context.TODO(),
	}
	cr := &moduleapi.Module{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test",
			Namespace:  "testns",
			Generation: 1,
		},
		Spec: moduleapi.ModuleSpec{
			ModuleName: string(moduleapi.CalicoModuleClass),
		},
	}

	uObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(cr)
	asserts.NoError(err)
	res, err := r.PreRemoveFinalizer(rctx, &unstructured.Unstructured{Object: uObj})
	asserts.NoError(err)
	asserts.False(res.Requeue)

	r.PostRemoveFinalizer(rctx, &unstructured.Unstructured{Object: uObj})
}

func (f FinalizerReconciler) Reconcile(reconcileContext controllerspi.ReconcileContext, u *unstructured.Unstructured) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (f FinalizerReconciler) GetReconcileObject() client.Object {
	return &moduleapi.Module{}
}
