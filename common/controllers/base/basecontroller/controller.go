// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package basecontroller

import (
	"context"
	"errors"
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/controllerspi"
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/watcher"
	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	vzstring "github.com/verrazzano/verrazzano-modules/common/pkg/string"
	vzlog "github.com/verrazzano/verrazzano-modules/common/pkg/vzlog"
	"go.uber.org/zap"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconcile the resource
// The controller-runtime will call this method repeatedly if the ctrl.Result.Requeue is true, or an error is returned
// This code will always return a nil error, and will set the ctrl.Result.Requeue to true (with a delay) if a requeue is needed.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Do some validation then get the GVK of the resource
	if r.Reconciler == nil {
		err := errors.New("Failed, Reconciler interface in ControllerConfig must be implemented")
		zap.S().Error(err)
		return util.NewRequeueWithShortDelay(), err
	}
	ro := r.GetReconcileObject()
	if ro == nil {
		err := errors.New("Failed, Reconciler.GetReconcileObject returns nil")
		zap.S().Error(err)
		return util.NewRequeueWithShortDelay(), err
	}
	gvk, _, err := r.Scheme.ObjectKinds(ro)
	if err != nil {
		zap.S().Errorf("Failed to get object GVK for %v: %v", r.GetReconcileObject(), err)
		return util.NewRequeueWithShortDelay(), nil
	}

	// Get the CR as unstructured
	cr := &unstructured.Unstructured{}
	cr.SetGroupVersionKind(gvk[0])
	if err := r.Get(ctx, req.NamespacedName, cr); err != nil {
		// If the resource is not found, that means all the finalizers have been removed,
		// and the Verrazzano resource has been deleted, so there is nothing left to do.
		if k8serrors.IsNotFound(err) {
			r.removeControllerResource(req.NamespacedName)
			return reconcile.Result{}, nil
		}
		zap.S().Errorf("Failed to fetch resource %v: %v", req.NamespacedName, err)
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
		return util.NewRequeueWithShortDelay(), nil
	}

	log.Progressf("Reconciling resource %v, GVK %v, generation %v", req.NamespacedName, gvk, cr.GetGeneration())

	// Create a new context for this reconcile loop
	rctx := controllerspi.ReconcileContext{
		Log:       vzlog.DefaultLogger(),
		ClientCtx: ctx,
	}

	// Handle finalizer
	if r.Finalizer != nil {
		// Make sure the ModuleCR has a finalizer
		if cr.GetDeletionTimestamp().IsZero() {
			if res := r.ensureFinalizer(log, cr); res.Requeue {
				return res, nil
			}
		} else {
			// ModuleCR is being deleted
			res, err := r.PreRemoveFinalizer(rctx, cr)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2, nil
			}

			if err := r.deleteFinalizer(log, cr); err != nil {
				return util.NewRequeueWithShortDelay(), nil
			}

			// all done, ModuleCR will be deleted from etcd
			log.Oncef("Successfully deleted resource %v, generation %v", req.NamespacedName, cr.GetGeneration())
			r.PostRemoveFinalizer(rctx, cr)
			r.removeControllerResource(req.NamespacedName)
			return ctrl.Result{}, nil
		}
	}

	if r.Watcher != nil {
		// Only keep track of resources if a watcher is used
		r.addControllerResource(req.NamespacedName)

		if err := r.initWatches(log, req.NamespacedName); err != nil {
			return util.NewRequeueWithShortDelay(), nil
		}
	}

	// Call the layered controller to reconcile.
	res, err := r.Reconciler.Reconcile(rctx, cr)
	if err != nil {
		return util.NewRequeueWithShortDelay(), nil
	}
	if util.ShouldRequeue(res) {
		return res, nil
	}

	// The resource has been reconciled.
	log.Infof("Successfully reconciled resource %v", req.NamespacedName)
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
			ResourceKind:               wds[i].WatchKind,
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

// ensureFinalizer ensures that a finalizer exists and updates the ModuleCR if it doesn't
func (r *Reconciler) ensureFinalizer(log vzlog.VerrazzanoLogger, u *unstructured.Unstructured) ctrl.Result {
	finalizerName := r.Finalizer.GetName()
	finalizers := u.GetFinalizers()
	if vzstring.SliceContainsString(finalizers, finalizerName) {
		return ctrl.Result{}
	}

	log.Debugf("Adding finalizer %s", finalizerName)
	finalizers = append(finalizers, finalizerName)
	u.SetFinalizers(finalizers)
	if err := r.Update(context.TODO(), u); err != nil {
		return util.NewRequeueWithShortDelay()
	}
	// Always requeue to make sure we don't reconcile until the status is finalizer is present
	return util.NewRequeueWithShortDelay()
}

// deleteFinalizer deletes the finalizer
func (r *Reconciler) deleteFinalizer(log vzlog.VerrazzanoLogger, u *unstructured.Unstructured) error {
	finalizerName := r.Finalizer.GetName()
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
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.controllerResources[nsn]
}

// GetControllerResources returns the list of controller resources
func (r *Reconciler) GetControllerResources() []types.NamespacedName {
	nsns := []types.NamespacedName{}
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for k := range r.controllerResources {
		nsns = append(nsns, k)
	}
	return nsns
}
