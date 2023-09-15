// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package basecontroller

import (
	"github.com/verrazzano/verrazzano-modules/pkg/controller/spi/controllerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	"k8s.io/apimachinery/pkg/types"
	controllerruntime "sigs.k8s.io/controller-runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// CreateControllerAndAddItToManager creates the base controller and adds it to the manager.
func CreateControllerAndAddItToManager(mgr controllerruntime.Manager, controllerConfig ControllerConfig) (*BaseReconciler, error) {

	r := BaseReconciler{
		Client:                  mgr.GetClient(),
		Scheme:                  mgr.GetScheme(),
		Cache:                   mgr.GetCache(),
		layeredControllerConfig: controllerConfig,
		watcherInitMap:          make(map[types.NamespacedName]bool),
		watchEvents:             make(map[types.NamespacedName]*controllerspi.WatchEvent),
	}

	// Create the controller and add it to the manager (Build does an implicit add)
	var err error
	if controllerConfig.EventFilter == nil {
		r.Controller, err = ctrl.NewControllerManagedBy(mgr).
			For(controllerConfig.Reconciler.GetReconcileObject()).
			WithOptions(controllerConfig.Options).
			Build(&r)
	} else {
		r.Controller, err = ctrl.NewControllerManagedBy(mgr).
			For(controllerConfig.Reconciler.GetReconcileObject()).
			WithEventFilter(r.createPredicateFilter(controllerConfig.EventFilter)).
			WithOptions(controllerConfig.Options).
			Build(&r)
	}

	if err != nil {
		return nil, vzlog.DefaultLogger().ErrorfNewErr("Failed calling SetupWithManager for Istio Gateway controller: %v", err)
	}
	return &r, nil
}

func (r *BaseReconciler) createPredicateFilter(filter controllerspi.EventFilter) predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return filter.HandlePredicateEvent(r.Client, e.Object)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return filter.HandlePredicateEvent(r.Client, e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return filter.HandlePredicateEvent(r.Client, e.ObjectOld)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return filter.HandlePredicateEvent(r.Client, e.Object)
		},
	}
}
