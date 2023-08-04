// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package basecontroller

import (
	"github.com/verrazzano/verrazzano-modules/pkg/controller/base/controllerspi"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Watch for a specific resource type
func (w *WatchContext) Watch() error {
	// The predicate functions are called to determine if the
	// controller reconcile loop needs to be called (predicate returns true)
	p := predicate.Funcs{
		// a watched resource just got created
		CreateFunc: func(e event.CreateEvent) bool {
			w.log.Infof("Watcher `create` occurred for watched resource %s/%s", e.Object.GetNamespace(), e.Object.GetName())
			return w.shouldReconcile(w.resourceBeingReconciled, e.Object, controllerspi.Created)
		},
		// a watched resource just got updated
		UpdateFunc: func(e event.UpdateEvent) bool {
			if e.ObjectOld == e.ObjectNew {
				return false
			}
			w.log.Infof("Watcher `update` event occurred for watched  resource %s/%s", e.ObjectNew.GetNamespace(), e.ObjectNew.GetName())
			return w.shouldReconcile(w.resourceBeingReconciled, e.ObjectNew, controllerspi.Updated)
		},
		// a watched resource just got deleted
		DeleteFunc: func(e event.DeleteEvent) bool {
			w.log.Infof("Watcher `delete` occurred for watched resource %s/%s", e.Object.GetNamespace(), e.Object.GetName())
			return w.shouldReconcile(w.resourceBeingReconciled, e.Object, controllerspi.Deleted)
		},
	}
	// return a Watch with the predicate that is called in the future when a resource
	// event occurs.  If the predicate returns true. then the reconciler loop will be called
	return w.controller.Watch(
		&w.watchDescriptor.WatchedResourceKind,
		w.createReconcileEventHandler(),
		p)
}

// createReconcileEventHandler creates an event handler that will get called
// when a watched event results in a true predicate.  The ModuleCR resource that this controller
// manages (meaning it exists) will be in the WatcherContext.
// A reconcile Request will be returned, causing the controller-runtime
// to call Reconcile for that resource.
func (w *WatchContext) createReconcileEventHandler() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(
		func(a client.Object) []reconcile.Request {
			requests := []reconcile.Request{}
			requests = append(requests, reconcile.Request{
				NamespacedName: w.resourceBeingReconciled})
			return requests
		})
}

// If the watched resource event should cause reconcile then return true
func (w *WatchContext) shouldReconcile(resourceBeingReconciled types.NamespacedName, watchedResource client.Object, ev controllerspi.WatchEvent) bool {
	if w.watchDescriptor.FuncShouldReconcile == nil {
		return false
	}
	return w.watchDescriptor.FuncShouldReconcile(resourceBeingReconciled, watchedResource, ev)
}
