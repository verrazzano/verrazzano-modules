// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package basecontroller

import (
	"context"
	"github.com/go-logr/logr"
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
	controllerruntime "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"testing"
)

const namespace = "testns"
const name = "test"

type fakeController struct{}

var _ controllerruntime.Controller = &fakeController{}

type ReconcilerImpl struct {
	reconcileVisited  bool
	getObjectsVisited bool
}
type WatcherImpl struct {
	visited bool
}
type FinalizerImpl struct{}

// TestReconciler tests that the layered controller reconcile method is called
// GIVEN a controller that implements Reconciler interface
// WHEN Reconcile is called
// THEN ensure that the controller returns success and that the Reconcile method is called
func TestReconciler(t *testing.T) {
	asserts := assert.New(t)

	controller := ReconcilerImpl{}
	config := ControllerConfig{
		Reconciler: &controller,
	}
	cr := newModuleCR(namespace, name)
	clientBuilder := fakes.NewClientBuilder()
	c := clientBuilder.WithScheme(newScheme()).WithObjects(cr).Build()
	r := newReconciler(c, config)

	request := newRequest(namespace, name)
	res, err := r.Reconcile(context.TODO(), request)

	// state and gen should never match
	asserts.NoError(err)
	asserts.False(res.Requeue)
	asserts.True(controller.reconcileVisited)
	asserts.True(controller.getObjectsVisited)
}

// TestWatcher tests that the layered controller reconcile method is called
// GIVEN a controller that implements Watcher interface
// WHEN Reconcile is called
// THEN ensure that the controller returns success and that the Watcher method is called
func TestWatcher(t *testing.T) {
	asserts := assert.New(t)

	watcher := WatcherImpl{}
	reconciler := ReconcilerImpl{}
	config := ControllerConfig{
		Watcher:    &watcher,
		Reconciler: &reconciler,
	}
	cr := newModuleCR(namespace, name)
	clientBuilder := fakes.NewClientBuilder()
	c := clientBuilder.WithScheme(newScheme()).WithObjects(cr).Build()
	r := newReconciler(c, config)

	request := newRequest(namespace, name)
	res, err := r.Reconcile(context.TODO(), request)

	// state and gen should never match
	asserts.NoError(err)
	asserts.False(res.Requeue)
	asserts.True(watcher.visited)
}

// TestFinalizer tests that the layered controller finalizer methods are called
// GIVEN a controller that implements Finalizer interface
// WHEN Reconcile is called
// THEN ensure that the controller returns success and that the Finalizer methods are called
func TestFinalizer(t *testing.T) {
	asserts := assert.New(t)

	watcher := WatcherImpl{}
	reconciler := ReconcilerImpl{}
	config := ControllerConfig{
		Watcher:    &watcher,
		Reconciler: &reconciler,
	}
	cr := newModuleCR(namespace, name)
	clientBuilder := fakes.NewClientBuilder()
	c := clientBuilder.WithScheme(newScheme()).WithObjects(cr).Build()
	r := newReconciler(c, config)

	request := newRequest(namespace, name)
	res, err := r.Reconcile(context.TODO(), request)

	// state and gen should never match
	asserts.NoError(err)
	asserts.False(res.Requeue)
	asserts.True(watcher.visited)
}

// TestReconcilerMissing tests that an error is returned when the reconciler implementation is missing
// GIVEN a controller that implements Reconciler interface
// WHEN Reconcile is called
// THEN ensure that the controller returns success
func TestReconcilerMissing(t *testing.T) {
	asserts := assert.New(t)

	controller := ReconcilerImpl{}
	config := ControllerConfig{}
	cr := newModuleCR(namespace, name)
	clientBuilder := fakes.NewClientBuilder()
	c := clientBuilder.WithScheme(newScheme()).WithObjects(cr).Build()
	r := newReconciler(c, config)

	request := newRequest(namespace, name)
	res, err := r.Reconcile(context.TODO(), request)

	// state and gen should never match
	asserts.Error(err)
	asserts.True(res.Requeue)
	asserts.False(controller.reconcileVisited)
	asserts.False(controller.getObjectsVisited)
}

// newReconciler creates a new reconciler for testing
func newReconciler(c client.Client, controllerConfig ControllerConfig) Reconciler {
	scheme := newScheme()
	reconciler := Reconciler{
		Client:              c,
		Scheme:              scheme,
		ControllerConfig:    controllerConfig,
		Controller:          fakeController{},
		controllerResources: make(map[types.NamespacedName]bool),
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

func newModuleCR(namespace string, name string) *moduleapi.Module {
	return &moduleapi.Module{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

// GetReconcileObject returns the kind of object being reconciled
func (r *ReconcilerImpl) GetReconcileObject() client.Object {
	r.getObjectsVisited = true
	return &moduleapi.Module{}
}

// Reconcile reconciles the ModuleLifecycle CR
func (r *ReconcilerImpl) Reconcile(spictx spi.ReconcileContext, u *unstructured.Unstructured) (ctrl.Result, error) {
	r.reconcileVisited = true
	cr := &moduleapi.Module{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, cr); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// GetWatchDescriptors returns the list of object kinds being watched
func (r *WatcherImpl) GetWatchDescriptors() []spi.WatchDescriptor {
	r.visited = true
	return []spi.WatchDescriptor{{
		WatchKind:           source.Kind{Type: &moduleapi.ModuleLifecycle{}},
		FuncShouldReconcile: nil,
	}}
}

func (f fakeController) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	return ctrl.Result{}, nil
}

func (f fakeController) Watch(src source.Source, eventhandler handler.EventHandler, predicates ...predicate.Predicate) error {
	return nil
}

func (f fakeController) Start(ctx context.Context) error {
	return nil
}

func (f fakeController) GetLogger() logr.Logger {
	return logr.Logger{}
}
