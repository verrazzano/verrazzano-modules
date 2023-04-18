// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package basecontroller

import (
	"context"
	"fmt"
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/spi"
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/watcher"
	vzctrl "github.com/verrazzano/verrazzano/pkg/controller"
	vzlog "github.com/verrazzano/verrazzano/pkg/log/vzlog"
	"go.uber.org/zap"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

// Reconcile the resource
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	cr := &unstructured.Unstructured{}
	gvk, _, err := r.Scheme.ObjectKinds(r.GetReconcileObject())
	if err != nil {
		zap.S().Errorf("Failed to get object GVK for %v: %v", r.GetReconcileObject(), err)
	}

	cr.SetGroupVersionKind(gvk[0])
	if err := r.Get(ctx, req.NamespacedName, cr); err != nil {
		// If the resource is not found, that means all the finalizers have been removed,
		// and the Verrazzano resource has been deleted, so there is nothing left to do.
		if k8serrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		zap.S().Errorf("Failed to fetch DNS resource: %v", err)
		return newRequeueWithDelay(), nil
	}

	log, err := vzlog.EnsureResourceLogger(&vzlog.ResourceConfig{
		Name:           cr.GetName(),
		Namespace:      cr.GetNamespace(),
		ID:             string(cr.GetUID()),
		Generation:     cr.GetGeneration(),
		ControllerName: r.Reconciler.GetReconcileObject().GetObjectKind().GroupVersionKind().Kind,
	})
	if err != nil {
		zap.S().Errorf("Failed to create controller logger for DNS controller", err)
	}

	log.Oncef("Reconciling resource %v, generation %v", req.NamespacedName, cr.GetGeneration())

	// Create a new context for this reconcile loop
	rctx := spi.ReconcileContext{
		Log:       vzlog.DefaultLogger(),
		ClientCtx: ctx,
	}

	// Handle finalizer
	if r.Finalizer != nil {
		if !cr.GetDeletionTimestamp().IsZero() {
			// Resource is getting deleted
			if err := r.deleteWatches(); err != nil {
				return newRequeueWithDelay(), nil
			}

			res, err := r.Cleanup(rctx, cr)
			if err != nil {
				return newRequeueWithDelay(), nil
			}
			if err != nil {
				return newRequeueWithDelay(), nil
			}
			if vzctrl.ShouldRequeue(res) {
				return res, nil
			}
			return ctrl.Result{}, nil
		}
		if err := r.ensureFinalizer(log); err != nil {
			return newRequeueWithDelay(), nil
		}
	}

	if r.Watcher != nil {
		if err := r.initWatches(log, cr.GetNamespace(), cr.GetName()); err != nil {
			return newRequeueWithDelay(), nil
		}
	}

	res, err := r.Reconciler.Reconcile(rctx, cr)
	if err != nil {
		return newRequeueWithDelay(), nil
	}
	if vzctrl.ShouldRequeue(res) {
		return res, nil
	}

	// The resource has been reconciled.
	log.Oncef("Successfully reconciled resource %v", req.NamespacedName)
	return ctrl.Result{}, nil
}

// Init the watches for this resource
func (r *Reconciler) initWatches(log vzlog.VerrazzanoLogger, namespace string, name string) error {
	if r.Watcher == nil {
		return nil
	}

	nsn := fmt.Sprintf("%s-%s", namespace, name)
	_, ok := r.watcherMap[nsn]
	if ok {
		return nil
	}

	// Get all the kinds of objects that need to be watched
	// For each object, create a watchContext and call the watcher to watch it
	watchKinds := r.Watcher.GetWatchedKinds()
	watchContexts := []watcher.WatchContext{}
	for i, _ := range watchKinds {
		w := watcher.WatchContext{
			Controller:      r.Controller,
			Log:             log,
			ResourceKind:    watchKinds[i].Kind,
			ShouldReconcile: watchKinds[i].FuncShouldReconcile,
		}
		err := w.Watch(namespace, name)
		if err != nil {
			return err
		}
		watchContexts = append(watchContexts, w)
	}

	r.watcherMap[nsn] = watchContexts
	return nil
}

// deleteWatches deletes the watches for this resource
func (r *Reconciler) deleteWatches() error {

	// TODO - must implement
	return nil
}

// ensureFinalizer ensures that a finalizer exists and updates the CR if it doesn't
func (r *Reconciler) ensureFinalizer(log vzlog.VerrazzanoLogger) error {

	// TODO - must implement
	return nil
}

// Create a new Result that will cause a reconciliation requeue after a short delay
func newRequeueWithDelay() ctrl.Result {
	return vzctrl.NewRequeueWithDelay(2, 3, time.Second)
}
