// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package watcher

import (
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/spi"
	"github.com/verrazzano/verrazzano/pkg/log/vzlog"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sync"
)

// WatchContext provides context to a watcher
type WatchContext struct {
	Controller      controller.Controller
	Log             vzlog.VerrazzanoLogger
	ResourceKind    source.Kind
	ShouldReconcile spi.FuncShouldReconcile

	touchedResources []*types.NamespacedName
	mutex            sync.RWMutex
}

// Watch for a specific resource type
func (w *WatchContext) Watch(reconciledResourceNamespace string, reconciledResourceName string) error {
	// The predicate functions are called to determine if the
	// controller reconcile loop needs to be called (predicate returns true)
	p := predicate.Funcs{
		// a watched resource just got created
		CreateFunc: func(e event.CreateEvent) bool {
			w.Log.Infof("Watcher `create` occurred for watched resource %s/%s", e.Object.GetNamespace(), e.Object.GetName())
			if !w.shouldReconcile(e.Object, spi.Created) {
				return false
			}
			w.addTouchedResource(e.Object)
			return true
		},
		// a watched resource just got updated
		UpdateFunc: func(e event.UpdateEvent) bool {
			if e.ObjectOld == e.ObjectNew {
				return false
			}
			w.Log.Infof("Watcher `update` event occurred for watched  resource %s/%s", e.ObjectNew.GetNamespace(), e.ObjectNew.GetName())
			if !w.shouldReconcile(e.ObjectNew, spi.Updated) {
				return false
			}
			w.addTouchedResource(e.ObjectNew)
			return true
		},
		// a watched resource just got deleted
		DeleteFunc: func(e event.DeleteEvent) bool {
			w.Log.Infof("Watcher `delete` occurred for watched resource %s/%s", e.Object.GetNamespace(), e.Object.GetName())
			if !w.shouldReconcile(e.Object, spi.Deleted) {
				return false
			}
			w.addTouchedResource(e.Object)
			return true
		},
	}
	// return a Watch with the predicate that is called in the future when a resource
	// event occurs.  If the predicate returns true then the reconcile loop will be called
	return w.Controller.Watch(
		&w.ResourceKind,
		w.createReconcileEventHandler(reconciledResourceNamespace, reconciledResourceName),
		p)
}

func (w *WatchContext) createReconcileEventHandler(namespace, name string) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(
		func(a client.Object) []reconcile.Request {
			return []reconcile.Request{
				{NamespacedName: types.NamespacedName{
					Namespace: namespace,
					Name:      name,
				}},
			}
		})
}

// If the touched resource should cause a reconcile then return true
func (w *WatchContext) shouldReconcile(watchedResource client.Object, ev spi.WatchEvent) bool {
	if w.ShouldReconcile == nil {
		return false
	}
	return w.ShouldReconcile(watchedResource, ev)
}

// Add touched resource to the map
func (w *WatchContext) addTouchedResource(watchedResource client.Object) bool {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.touchedResources = append(w.touchedResources,
		&types.NamespacedName{Namespace: watchedResource.GetNamespace(), Name: watchedResource.GetName()})
	return true
}

// GetTouchedNames gets the names of resources that were touched
func (w *WatchContext) GetTouchedNames() []*types.NamespacedName {
	names := []*types.NamespacedName{}
	if len(w.touchedResources) > 0 {
		w.mutex.Lock()
		defer w.mutex.Unlock()
		for i := range w.touchedResources {
			names = append(names, w.touchedResources[i])
		}
	}
	return names
}
