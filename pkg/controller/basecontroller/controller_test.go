// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package basecontroller

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/result"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/spi/controllerspi"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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
	"sync"
	"testing"
)

const namespace = "testns"
const name = "test"
const finalizerName = "fimalizer"

type fakeController struct{}

var _ controllerruntime.Controller = &fakeController{}

type ReconcilerImpl struct {
	reconcileCalled bool
	getObjectCalled bool
	returnNilObject bool
}

type WatcherImpl struct {
	called bool
}
type FinalizerImpl struct {
	getNameCalled     bool
	preCleanupCalled  bool
	postCleanupCalled bool
}

// TestConcurrentReconcile tests that the layered controller interfaces are called
// GIVEN a controller that implements the controllers spi interfaces
// WHEN Reconcile is called
// THEN ensure that the controller returns success and that the interface methods are all called
func TestConcurrentReconcile(t *testing.T) {
	asserts := assert.New(t)

	const threadCount = 100
	var wg sync.WaitGroup

	// Load CRs
	controller := ReconcilerImpl{}
	watcher := WatcherImpl{}
	finalizer := FinalizerImpl{}
	config := ControllerConfig{
		Reconciler: &controller,
		Watcher:    &watcher,
		Finalizer:  &finalizer,
	}
	clientBuilder := fakes.NewClientBuilder().WithScheme(newScheme())
	for i := 1; i <= threadCount; i++ {
		crName := fmt.Sprintf("%s-%v", name, i)
		cr := newModuleCR(namespace, crName)
		addFinalizer(cr)
		clientBuilder = clientBuilder.WithObjects(cr)
	}
	c := clientBuilder.Build()
	r := newReconciler(c, config)

	// Reconcile CRs in parallel
	for i := 1; i <= threadCount; i++ {
		wg.Add(1)
		go func(y int) {
			defer wg.Done()

			crName := fmt.Sprintf("%s-%v", name, y)
			request := newRequest(namespace, crName)
			res, err := r.Reconcile(context.TODO(), request)

			// state and gen should never match
			asserts.NoError(err)
			asserts.False(res.Requeue)
			asserts.True(controller.reconcileCalled)
			asserts.True(controller.getObjectCalled)
			asserts.True(watcher.called)
			asserts.True(finalizer.getNameCalled)
			asserts.False(finalizer.preCleanupCalled)
			asserts.False(finalizer.postCleanupCalled)
		}(i)
	}
	wg.Wait()
}

// TestReconciler tests that the layered controller Reconcile interface works alone
// GIVEN a controller that implements BaseReconciler interface
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
	asserts.True(controller.reconcileCalled)
	asserts.True(controller.getObjectCalled)
}

// TestWatcher tests that the layered controller reconcile method is called
// GIVEN a controller that implements Watcher interface
// WHEN Reconcile is called
// THEN ensure that the controller returns success and that the Watcher method is called
func TestWatcher(t *testing.T) {
	asserts := assert.New(t)

	reconciler := ReconcilerImpl{}
	watcher := WatcherImpl{}
	config := ControllerConfig{
		Reconciler: &reconciler,
		Watcher:    &watcher,
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
	asserts.True(watcher.called)
}

// TestEnsureFinalizer tests that a finalizer is added to the ModuleCR
// GIVEN a controller that implements Finalizer interface
// WHEN Reconcile is called
// THEN ensure that the ModuleCR is updated with the finalizer
func TestEnsureFinalizer(t *testing.T) {
	asserts := assert.New(t)

	reconciler := ReconcilerImpl{}
	finalizer := FinalizerImpl{}
	config := ControllerConfig{
		Reconciler: &reconciler,
		Finalizer:  &finalizer,
	}
	cr := newModuleCR(namespace, name)
	clientBuilder := fakes.NewClientBuilder()
	c := clientBuilder.WithScheme(newScheme()).WithObjects(cr).Build()
	r := newReconciler(c, config)

	request := newRequest(namespace, name)
	res, err := r.Reconcile(context.TODO(), request)

	// state and gen should never match
	asserts.NoError(err)
	asserts.True(res.Requeue)
	asserts.True(finalizer.getNameCalled)
	asserts.False(finalizer.preCleanupCalled)
	asserts.False(finalizer.postCleanupCalled)

	updatedCR := moduleapi.Module{}
	err = r.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: name}, &updatedCR)
	asserts.NoError(err)
	asserts.Len(updatedCR.Finalizers, 1)
}

// TestFinalizerAlreadyExists tests that a finalizer is mot added to the ModuleCR if it exists
// GIVEN a controller that implements Finalizer interface
// WHEN Reconcile is called and the finalizer exists in the ModuleCR
// THEN ensure that the ModuleCR is not updated with another finalizer
func TestFinalizerAlreadyExists(t *testing.T) {
	asserts := assert.New(t)

	reconciler := ReconcilerImpl{}
	finalizer := FinalizerImpl{}
	config := ControllerConfig{
		Reconciler: &reconciler,
		Finalizer:  &finalizer,
	}
	cr := newModuleCR(namespace, name)
	addFinalizer(cr)
	clientBuilder := fakes.NewClientBuilder()
	c := clientBuilder.WithScheme(newScheme()).WithObjects(cr).Build()
	r := newReconciler(c, config)

	request := newRequest(namespace, name)
	res, err := r.Reconcile(context.TODO(), request)

	// state and gen should never match
	asserts.NoError(err)
	asserts.False(res.Requeue)
	asserts.True(finalizer.getNameCalled)
	asserts.False(finalizer.preCleanupCalled)
	asserts.False(finalizer.postCleanupCalled)

	updatedCR := moduleapi.Module{}
	err = r.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: name}, &updatedCR)
	asserts.NoError(err)
	asserts.Len(updatedCR.Finalizers, 1)
}

// TestDelete tests that the layered controller finalizer methods are called
// GIVEN a controller that implements Finalizer interface
// WHEN Reconcile is called
// THEN ensure that the controller returns success and that the Finalizer methods are called
func TestDelete(t *testing.T) {
	asserts := assert.New(t)

	reconciler := ReconcilerImpl{}
	finalizer := FinalizerImpl{}
	config := ControllerConfig{
		Reconciler: &reconciler,
		Finalizer:  &finalizer,
	}
	cr := newModuleCR(namespace, name)
	addFinalizer(cr)
	addDeletionTimestamp(cr)
	clientBuilder := fakes.NewClientBuilder()
	c := clientBuilder.WithScheme(newScheme()).WithObjects(cr).Build()
	r := newReconciler(c, config)

	request := newRequest(namespace, name)
	res, err := r.Reconcile(context.TODO(), request)

	// state and gen should never match
	asserts.NoError(err)
	asserts.False(res.Requeue)
	asserts.True(finalizer.getNameCalled)
	asserts.True(finalizer.preCleanupCalled)
	asserts.True(finalizer.postCleanupCalled)

	updatedCR := moduleapi.Module{}
	err = r.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: name}, &updatedCR)
	asserts.True(errors.IsNotFound(err))
}

// TestReconcilerMissing tests that an error is returned when the reconciler implementation is missing
// GIVEN a controller that implements BaseReconciler interface
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
	asserts.False(controller.reconcileCalled)
	asserts.False(controller.getObjectCalled)
}

// TestReconcilerGetObjectMissing tests that an error is returned
// GIVEN a controller that implements BaseReconciler interface
// WHEN Reconcile is called and GetReconcileObject returns nil
// THEN ensure that the controller returns and error
func TestReconcilerGetObjectMissing(t *testing.T) {
	asserts := assert.New(t)

	reconciler := ReconcilerImpl{returnNilObject: true}
	config := ControllerConfig{
		Reconciler: &reconciler,
	}
	cr := newModuleCR(namespace, name)
	clientBuilder := fakes.NewClientBuilder()
	c := clientBuilder.WithScheme(newScheme()).WithObjects(cr).Build()
	r := newReconciler(c, config)

	request := newRequest(namespace, name)
	res, err := r.Reconcile(context.TODO(), request)

	// state and gen should never match
	asserts.Error(err)
	asserts.True(res.Requeue)
	asserts.False(reconciler.reconcileCalled)
	asserts.True(reconciler.getObjectCalled)
}

// TestNotFound tests that the controller handles not found
// GIVEN a controller that implements BaseReconciler interface
// WHEN Reconcile is called
// THEN ensure that the controller returns success if ModuleCR doesn't exist
func TestNotFound(t *testing.T) {
	asserts := assert.New(t)

	reconciler := ReconcilerImpl{}
	config := ControllerConfig{
		Reconciler: &reconciler,
	}
	clientBuilder := fakes.NewClientBuilder()
	c := clientBuilder.WithScheme(newScheme()).Build()
	r := newReconciler(c, config)

	request := newRequest(namespace, name)
	res, err := r.Reconcile(context.TODO(), request)

	// state and gen should never match
	asserts.NoError(err)
	asserts.False(res.Requeue)
}

// newReconciler creates a new reconciler for testing
func newReconciler(c client.Client, controllerConfig ControllerConfig) *BaseReconciler {
	scheme := newScheme()
	reconciler := BaseReconciler{
		Client:                  c,
		Scheme:                  scheme,
		layeredControllerConfig: controllerConfig,
		Controller:              fakeController{},
		watcherInitMap:          make(map[types.NamespacedName]bool),
		watchEvents:             make(map[types.NamespacedName]*controllerspi.WatchEvent),
	}
	return &reconciler
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
	m := &moduleapi.Module{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	return m
}

func addDeletionTimestamp(m *moduleapi.Module) {
	t := metav1.Now()
	m.ObjectMeta.DeletionTimestamp = &t
}

func addFinalizer(m *moduleapi.Module) {
	m.ObjectMeta.Finalizers = []string{finalizerName}
}

// GetReconcileObject returns the kind of object being reconciled
func (r *ReconcilerImpl) GetReconcileObject() client.Object {
	r.getObjectCalled = true
	if r.returnNilObject {
		return nil
	}
	return &moduleapi.Module{}
}

// Reconcile reconciles the ModuleAction ModuleCR
func (r *ReconcilerImpl) Reconcile(spictx controllerspi.ReconcileContext, u *unstructured.Unstructured) result.Result {
	r.reconcileCalled = true
	cr := &moduleapi.Module{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, cr); err != nil {
		return result.NewResult()
	}
	return result.NewResult()
}

// GetWatchDescriptors returns the list of object kinds being watched
func (r *WatcherImpl) GetWatchDescriptors() []controllerspi.WatchDescriptor {
	r.called = true
	return []controllerspi.WatchDescriptor{{
		WatchedResourceKind: &moduleapi.Module{},
		FuncShouldReconcile: nil,
	}}
}

func (f *FinalizerImpl) GetName() string {
	f.getNameCalled = true
	return finalizerName
}

func (f *FinalizerImpl) PreRemoveFinalizer(reconcileContext controllerspi.ReconcileContext, u *unstructured.Unstructured) result.Result {
	f.preCleanupCalled = true
	return result.NewResult()
}

func (f *FinalizerImpl) PostRemoveFinalizer(reconcileContext controllerspi.ReconcileContext, u *unstructured.Unstructured) {
	f.postCleanupCalled = true
}

func (f fakeController) Reconcile(ctx context.Context, request reconcile.Request) (ctrl.Result, error) {
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
