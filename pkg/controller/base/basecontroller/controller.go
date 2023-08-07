// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package basecontroller

import (
	"context"
	"errors"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/base/controllerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/result"
	vzstring "github.com/verrazzano/verrazzano-modules/pkg/string"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	"go.uber.org/zap"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

// Reconcile the resource
// The controller-runtime will call this method repeatedly if the ctrl.Result.Requeue is true, or an error is returned
// This code will always return a nil error, and will set the ctrl.Result.Requeue to true (with a delay) if a requeue is needed.
func (r *BaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Do some validation then get the GVK of the resource
	if r.layeredControllerConfig.Reconciler == nil {
		err := errors.New("Failed, BaseReconciler interface in ControllerConfig must be implemented")
		zap.S().Error(err)
		return result.NewResultShortRequeueDelay().GetCtrlRuntimeResult(), err
	}
	ro := r.layeredControllerConfig.GetReconcileObject()
	if ro == nil {
		err := errors.New("Failed, BaseReconciler.GetReconcileObject returns nil")
		zap.S().Error(err)
		return result.NewResultShortRequeueDelay().GetCtrlRuntimeResult(), err
	}
	gvk, _, err := r.Scheme.ObjectKinds(ro)
	if err != nil {
		zap.S().Errorf("Failed to get object GVK for %v: %v", r.layeredControllerConfig.GetReconcileObject(), err)
		return result.NewResultShortRequeueDelay().GetCtrlRuntimeResult(), nil
	}

	// Get the CR as unstructured
	cr := &unstructured.Unstructured{}
	cr.SetGroupVersionKind(gvk[0])
	if err := r.Get(ctx, req.NamespacedName, cr); err != nil {
		// If the resource is not found, that means all the finalizers have been removed,
		// and the  resource has been deleted, so there is nothing left to do.
		if k8serrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		zap.S().Errorf("Failed to fetch resource %v: %v", req.NamespacedName, err)
		return result.NewResultShortRequeueDelay().GetCtrlRuntimeResult(), nil
	}

	log, err := vzlog.EnsureResourceLogger(&vzlog.ResourceConfig{
		Name:           cr.GetName(),
		Namespace:      cr.GetNamespace(),
		ID:             string(cr.GetUID()),
		Generation:     cr.GetGeneration(),
		ControllerName: r.layeredControllerConfig.Reconciler.GetReconcileObject().GetObjectKind().GroupVersionKind().Kind,
	})
	if err != nil {
		zap.S().Errorf("Failed to create controller logger for DNS controller", err)
		return result.NewResultShortRequeueDelay().GetCtrlRuntimeResult(), nil
	}

	log.Debugf("Reconciling resource %v, GVK %v, generation %v", req.NamespacedName, gvk, cr.GetGeneration())

	// Create a new context for this reconcile loop
	rctx := controllerspi.ReconcileContext{
		Log:       log,
		ClientCtx: ctx,
	}

	// Handle finalizer
	if r.layeredControllerConfig.Finalizer != nil {
		// Make sure the ModuleCR has a finalizer
		if cr.GetDeletionTimestamp().IsZero() {
			if res := r.ensureFinalizer(log, cr); res.ShouldRequeue() {
				return res.GetCtrlRuntimeResult(), nil
			}
		} else {
			// ModuleCR is being deleted
			res := r.layeredControllerConfig.PreRemoveFinalizer(rctx, cr)
			if res.ShouldRequeue() {
				return res.GetCtrlRuntimeResult(), nil
			}

			if err := r.deleteFinalizer(log, cr); err != nil {
				return result.NewResultShortRequeueDelay().GetCtrlRuntimeResult(), nil
			}

			// all done, ModuleCR will be deleted from etcd
			log.Oncef("Successfully deleted resource %v, generation %v", req.NamespacedName, cr.GetGeneration())
			r.layeredControllerConfig.PostRemoveFinalizer(rctx, cr)
			return ctrl.Result{}, nil
		}
	}

	if r.layeredControllerConfig.Watcher != nil {
		if err := r.initWatches(log, req.NamespacedName); err != nil {
			return result.NewResultShortRequeueDelay().GetCtrlRuntimeResult(), nil
		}
	}

	// Call the layered controller to reconcile.
	res := r.layeredControllerConfig.Reconciler.Reconcile(rctx, cr)
	if err != nil {
		return result.NewResultShortRequeueDelay().GetCtrlRuntimeResult(), nil
	}
	if res.ShouldRequeue() {
		return res.GetCtrlRuntimeResult(), nil
	}

	// The resource has been reconciled.
	log.Debugf("Successfully reconciled resource %v", req.NamespacedName)
	return ctrl.Result{}, nil
}

// Init the watches for this resource
func (r *BaseReconciler) initWatches(log vzlog.VerrazzanoLogger, resourceNSN types.NamespacedName) error {
	r.watchMutex.Lock()
	defer r.watchMutex.Unlock()
	if _, ok := r.watcherInitMap[resourceNSN]; ok {
		return nil
	}

	// For each object being watched, create a watchContext and call the watcher to watch it
	wds := r.layeredControllerConfig.Watcher.GetWatchDescriptors()
	for i := range wds {
		w := &WatchContext{
			controller:              r.Controller,
			reconciler:              r,
			log:                     log,
			watchDescriptor:         wds[i],
			resourceBeingReconciled: resourceNSN,
		}
		err := w.Watch()
		if err != nil {
			return err
		}
		r.watchContexts = append(r.watchContexts, w)
	}

	r.watcherInitMap[resourceNSN] = true
	return nil
}

// ensureFinalizer ensures that a finalizer exists and updates the ModuleCR if it doesn't
func (r *BaseReconciler) ensureFinalizer(log vzlog.VerrazzanoLogger, u *unstructured.Unstructured) result.Result {
	finalizerName := r.layeredControllerConfig.Finalizer.GetName()
	finalizers := u.GetFinalizers()
	if vzstring.SliceContainsString(finalizers, finalizerName) {
		return result.NewResult()
	}

	log.Debugf("Adding finalizer %s", finalizerName)
	finalizers = append(finalizers, finalizerName)
	u.SetFinalizers(finalizers)
	if err := r.Update(context.TODO(), u); err != nil {
		return result.NewResultShortRequeueDelay()
	}
	// Always requeue to make sure we don't reconcile until the status is finalizer is present
	return result.NewResultShortRequeueDelay()
}

// deleteFinalizer deletes the finalizer
func (r *BaseReconciler) deleteFinalizer(log vzlog.VerrazzanoLogger, u *unstructured.Unstructured) error {
	finalizerName := r.layeredControllerConfig.Finalizer.GetName()
	finalizers := u.GetFinalizers()
	if !vzstring.SliceContainsString(finalizers, finalizerName) {
		return nil
	}
	log.Debugf("Removing finalizer %s", finalizerName)
	finalizers = vzstring.RemoveStringFromSlice(u.GetFinalizers(), finalizerName)
	u.SetFinalizers(finalizers)
	if err := r.Update(context.TODO(), u); err != nil {
		return err
	}

	return nil
}

// Update the map entry for the resource with the current time
func (r *BaseReconciler) updateWatchTimestamp(nsn types.NamespacedName) {
	r.watchMutex.Lock()
	defer r.watchMutex.Unlock()
	r.watchEventTimestampMap[nsn] = time.Now()
}

// GetWatchTimestamp gets the map entry for the resource
func (r *BaseReconciler) GetWatchTimestamp(nsn types.NamespacedName) *time.Time {
	r.watchMutex.Lock()
	defer r.watchMutex.Unlock()
	t, ok := r.watchEventTimestampMap[nsn]
	if !ok {
		return nil
	}
	return &t
}
