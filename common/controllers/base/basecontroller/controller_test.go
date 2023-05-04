// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package basecontroller

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/spi"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakes "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"testing"
)

type ReconcilerImpl struct{}
type WatcherImpl struct{}
type FinalizerImpl struct{}

// TestEmptyControllerContig tests that the layered controller reconcile method is called
// GIVEN a Reconciler
// WHEN Reconcile is called
// THEN ensure that the layered controller returns success
func TestEmptyControllerContig(t *testing.T) {
	asserts := assert.New(t)

	const namespace = "testns"
	const name = "test"

	reconcilerImpl := ReconcilerImpl{}
	config := ControllerConfig{
		Reconciler: reconcilerImpl,
	}
	cr := &moduleapi.Module{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	clientBuilder := fakes.NewClientBuilder()
	c := clientBuilder.WithScheme(newScheme()).WithObjects(cr).Build()
	r := newReconciler(c, config)

	request := newRequest(namespace, name)
	res, err := r.Reconcile(context.TODO(), request)

	// state and gen should never match
	asserts.NoError(err)
	asserts.False(res.Requeue)
}

// newReconciler creates a new reconciler for testing
func newReconciler(c client.Client, controllerConfig ControllerConfig) Reconciler {
	scheme := newScheme()
	reconciler := Reconciler{
		Client:           c,
		Scheme:           scheme,
		ControllerConfig: controllerConfig,
	}
	return reconciler
}

// newScheme creates a new scheme that includes this package's object to use for testing
func newScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = moduleapi.AddToScheme(scheme)
	return scheme
}

func newRequest(namespace string, name string) ctrl.Request {
	return ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: namespace,
			Name:      name}}
}

// GetReconcileObject returns the kind of object being reconciled
func (r ReconcilerImpl) GetReconcileObject() client.Object {
	return &moduleapi.Module{}
}

// Reconcile reconciles the ModuleLifecycle CR
func (r ReconcilerImpl) Reconcile(spictx spi.ReconcileContext, u *unstructured.Unstructured) (ctrl.Result, error) {
	cr := &moduleapi.Module{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, cr); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// GetWatchDescriptors returns the list of object kinds being watched
func (r WatcherImpl) GetWatchDescriptors() []spi.WatchDescriptor {
	return []spi.WatchDescriptor{{
		WatchKind:           source.Kind{Type: &moduleapi.ModuleLifecycle{}},
		FuncShouldReconcile: nil,
	}}
}
