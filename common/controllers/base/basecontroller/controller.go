// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package basecontroller

import (
	"context"
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/spi"
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/watcher"
	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	vzctrl "github.com/verrazzano/verrazzano/pkg/controller"
	vzlog "github.com/verrazzano/verrazzano/pkg/log/vzlog"
	vzstring "github.com/verrazzano/verrazzano/pkg/string"
	"go.uber.org/zap"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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
			r.removeControllerResource(req.NamespacedName)
			return reconcile.Result{}, nil
		}
		zap.S().Errorf("Failed to fetch DNS resource: %v", err)
		return util.NewRequeueWithShortDelay(), nil
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
		// Make sure the CR has a finalizer, if it is not being deleted
		if cr.GetDeletionTimestamp().IsZero() {
			if err := r.ensureFinalizer(log, cr); err != nil {
				return util.NewRequeueWithShortDelay(), nil
			}
		} else {
			res, err := r.Cleanup(rctx, cr)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2, nil
			}

			if err := r.deleteFinalizer(log, cr); err != nil {
				return util.NewRequeueWithShortDelay(), nil
			}

			// all done, CR will be deleted from etcd
			log.Oncef("Successfully deleted resource %v, generation %v", req.NamespacedName, cr.GetGeneration())
			r.removeControllerResource(req.NamespacedName)
			return ctrl.Result{}, nil
		}
	}

	if r.Watcher != nil {
		r.addControllerResource(req.NamespacedName)
		if err := r.initWatches(log, req.NamespacedName); err != nil {
			return util.NewRequeueWithShortDelay(), nil
		}
	}

	res, err := r.Reconciler.Reconcile(rctx, cr)
	if err != nil {
		return util.NewRequeueWithShortDelay(), nil
	}
	if vzctrl.ShouldRequeue(res) {
		return res, nil
	}

	// The resource has been reconciled.
	log.Oncef("Successfully reconciled resource %v", req.NamespacedName)
	return ctrl.Result{}, nil
}

// Init the watches for this resource
func (r *Reconciler) initWatches(log vzlog.VerrazzanoLogger, nsn types.NamespacedName) error {
	if r.watchersInitialized {
		return nil
	}

	// Get all the kinds of objects that need to be watched
	// For each object, create a watchContext and call the watcher to watch it
	wds := r.Watcher.GetWatchDescriptors()
	for i := range wds {
		w := &watcher.WatchContext{
			Controller:                 r.Controller,
			Log:                        log,
			ResourceKind:               wds[i].Kind,
			ShouldReconcile:            wds[i].FuncShouldReconcile,
			FuncGetControllerResources: r.GetControllerResources,
		}
		err := w.Watch()
		if err != nil {
			return err
		}
		r.watchContexts = append(r.watchContexts, w)
	}

	r.watchersInitialized = true
	return nil
}

// ensureFinalizer ensures that a finalizer exists and updates the CR if it doesn't
func (r *Reconciler) ensureFinalizer(log vzlog.VerrazzanoLogger, u *unstructured.Unstructured) error {
	finalizerName := r.Finalizer.GetName()
	finalizers := u.GetFinalizers()
	if vzstring.SliceContainsString(finalizers, finalizerName) {
		return nil
	}

	log.Oncef("Adding finalizer %s", finalizerName)
	finalizers = append(finalizers, finalizerName)
	u.SetFinalizers(finalizers)
	if err := r.Update(context.TODO(), u); err != nil {
		return err
	}

	return nil
}

// deleteFinalizer deletes the finalizer
func (r *Reconciler) deleteFinalizer(log vzlog.VerrazzanoLogger, u *unstructured.Unstructured) error {
	finalizerName := r.Finalizer.GetName()
	finalizers := u.GetFinalizers()
	if !vzstring.SliceContainsString(finalizers, finalizerName) {
		return nil
	}
	log.Oncef("Removing finalizer %s", finalizerName)
	finalizers = vzstring.RemoveStringFromSlice(u.GetFinalizers(), finalizerName)
	u.SetFinalizers(finalizers)
	if err := r.Update(context.TODO(), u); err != nil {
		return err
	}

	return nil
}

// removeControllerResource removes a controller resource from the set
func (r *Reconciler) removeControllerResource(nsn types.NamespacedName) {
	if !r.controllerResourceExists(nsn) {
		return
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.controllerResources, nsn)
}

// addControllerResource adds a controller resource to the set
func (r *Reconciler) addControllerResource(nsn types.NamespacedName) {
	if r.controllerResourceExists(nsn) {
		return
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.controllerResources[nsn] = true
}

// controllerResourceExists returns true if the resource is in the set
func (r *Reconciler) controllerResourceExists(nsn types.NamespacedName) bool {
	if r.controllerResources == nil {
		return false
	}
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.controllerResources[nsn]
}

// GetControllerResources returns the list of controller resources
func (r *Reconciler) GetControllerResources() []types.NamespacedName {
	nsns := []types.NamespacedName{}
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	for k := range r.controllerResources {
		nsns = append(nsns, k)
	}
	return nsns
}
