// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package basecontroller

import (
	"github.com/verrazzano/verrazzano-modules/pkg/controller/base/controllerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// FuncGetControllerResources is a type of function that resources the CRs that exist which are managed bu the controller
type FuncGetControllerResources func() []types.NamespacedName

// WatchContext provides context to a watcher
type WatchContext struct {
	Controller      controller.Controller
	Log             vzlog.VerrazzanoLogger
	ResourceKind    source.Kind
	ShouldReconcile controllerspi.FuncShouldReconcile
	FuncGetControllerResources
}

// Watch for a specific resource type
func (w *WatchContext) Watch() error {
	// The predicate functions are called to determine if the
	// controller reconcile loop needs to be called (predicate returns true)
	p := predicate.Funcs{
		// a watched resource just got created
		CreateFunc: func(e event.CreateEvent) bool {
			w.Log.Infof("Watcher `create` occurred for watched resource %s/%s", e.Object.GetNamespace(), e.Object.GetName())
			return w.shouldReconcile(e.Object, controllerspi.Created)
		},
		// a watched resource just got updated
		UpdateFunc: func(e event.UpdateEvent) bool {
			if e.ObjectOld == e.ObjectNew {
				return false
			}
			w.Log.Infof("Watcher `update` event occurred for watched  resource %s/%s", e.ObjectNew.GetNamespace(), e.ObjectNew.GetName())
			return w.shouldReconcile(e.ObjectNew, controllerspi.Updated)
		},
		// a watched resource just got deleted
		DeleteFunc: func(e event.DeleteEvent) bool {
			w.Log.Infof("Watcher `delete` occurred for watched resource %s/%s", e.Object.GetNamespace(), e.Object.GetName())
			return w.shouldReconcile(e.Object, controllerspi.Deleted)
		},
	}
	// return a Watch with the predicate that is called in the future when a resource
	// event occurs.  If the predicate returns true. then the reconciler loop will be called
	return w.Controller.Watch(
		&w.ResourceKind,
		w.createReconcileEventHandler(),
		p)
}

// createReconcileEventHandler creates an event handler that will get called
// when a watched event results in a true predicate.  Each ModuleCR resource that this controller
// manages (meaning it exists) will be in the WatcherContext.reconcileResources list.
// A reconcile.Request will be returned for each resource, causing the controller-runtime
// to call Reconcile for that resource.
func (w *WatchContext) createReconcileEventHandler() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(
		func(a client.Object) []reconcile.Request {
			requests := []reconcile.Request{}
			resources := w.FuncGetControllerResources()
			for i := range resources {
				requests = append(requests, reconcile.Request{
					NamespacedName: resources[i]})
			}
			return requests
		})
}

// If the watched resource event should cause reconcile then return true
func (w *WatchContext) shouldReconcile(watchedResource client.Object, ev controllerspi.WatchEvent) bool {
	if w.ShouldReconcile == nil {
		return false
	}
	return w.ShouldReconcile(watchedResource, ev)
}
