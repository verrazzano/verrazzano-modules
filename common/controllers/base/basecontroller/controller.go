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
		// If the resource is not found, that means all of the finalizers have been removed,
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
		ControllerName: r.DescribeController.GetReconcileObject().GetObjectKind().GroupVersionKind().Kind,
	})
	if err != nil {
		zap.S().Errorf("Failed to create controller logger for DNS controller", err)
	}

	// Create a new context for this reconcile loop
	rctx := spi.ReconcileContext{
		Log:       vzlog.DefaultLogger(),
		ClientCtx: ctx,
	}

	rctx.Log.Oncef("Reconciling Verrazzano resource %v", req.NamespacedName)
	if !cr.GetDeletionTimestamp().IsZero() {
		// TODO - handle finalizer - use FinalizerController interface
		return reconcile.Result{}, nil
	}

	if err := r.initWatches(log, cr.GetNamespace(), cr.GetName()); err != nil {
		return newRequeueWithDelay(), nil
	}

	if err = r.ReconcileController.Reconcile(rctx, cr); err != nil {
		return newRequeueWithDelay(), nil
	}

	// The resource has been reconciled.
	log.Oncef("Successfully reconciled Gateway resource %v", req.NamespacedName)
	return ctrl.Result{}, nil
}

// Create a new Result that will cause a reconcile requeue after a short delay
func newRequeueWithDelay() ctrl.Result {
	return vzctrl.NewRequeueWithDelay(1, 2, time.Second)
}

// Init the watch for this resource
func (r *Reconciler) initWatches(log vzlog.VerrazzanoLogger, namespace string, name string) error {
	if r.WatchController == nil {
		return nil
	}

	nsn := fmt.Sprintf("%s-%s", namespace, name)
	_, ok := r.watcherMap[nsn]
	if ok {
		return nil
	}

	// Get all the kinds of objects that need to be watched
	// For each object, create a watchContext and call the watcher to watch it
	watchKinds := r.WatchController.GetWatchedKinds()
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
