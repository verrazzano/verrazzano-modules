// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package basecontroller

import (
	"context"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/base/controllerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	controllerruntime "sigs.k8s.io/controller-runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sync"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// ControllerConfig specifies the config of the controller using this base controller
type ControllerConfig struct {
	controllerspi.Finalizer
	controllerspi.Reconciler
	controllerspi.Watcher
}

// Reconciler contains data needed to reconcile a DNS object.
type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Controller controller.Controller
	ControllerConfig
	LifecycleClass      moduleapi.ModuleClassType
	watchersInitialized bool
	watchContexts       []*WatchContext

	// controllerResources contains a set of CRs for this controller that exist.
	// It is important that resources get added to this set during the base controller reconcile loop, as
	// well as removed when the resource is deleted.
	controllerResources map[types.NamespacedName]bool
	mutex               sync.RWMutex
}

// CreateControllerAndAddItToManager creates the base controller and adds it to the manager.
func CreateControllerAndAddItToManager(mgr controllerruntime.Manager, controllerConfig ControllerConfig, class moduleapi.ModuleClassType) (*Reconciler, error) {
	r := Reconciler{
		Client:              mgr.GetClient(),
		Scheme:              mgr.GetScheme(),
		ControllerConfig:    controllerConfig,
		LifecycleClass:      class,
		controllerResources: make(map[types.NamespacedName]bool),
		mutex:               sync.RWMutex{},
	}

	// Create the controller and add it to the manager (Build does an implicit add)
	var err error
	if class == "" {
		r.Controller, err = ctrl.NewControllerManagedBy(mgr).
			For(controllerConfig.Reconciler.GetReconcileObject()).
			Build(&r)
	} else {
		r.Controller, err = ctrl.NewControllerManagedBy(mgr).
			For(controllerConfig.Reconciler.GetReconcileObject()).
			WithEventFilter(r.createPredicateFilter()).
			Build(&r)
	}

	if err != nil {
		return nil, vzlog.DefaultLogger().ErrorfNewErr("Failed calling SetupWithManager for Istio Gateway controller: %v", err)
	}
	return &r, nil
}

func (r *Reconciler) createPredicateFilter() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return r.handlesEvent(e.Object)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return r.handlesEvent(e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return r.handlesEvent(e.ObjectOld)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return r.handlesEvent(e.Object)
		},
	}
}

func (r *Reconciler) handlesEvent(object client.Object) bool {
	mlc := moduleapi.ModuleAction{}
	objectkey := client.ObjectKeyFromObject(object)
	if err := r.Get(context.TODO(), objectkey, &mlc); err != nil {
		return false
	}
	return mlc.Spec.ModuleClassName == r.LifecycleClass
}
