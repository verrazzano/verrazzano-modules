// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package basecontroller

import (
	"context"
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/spi"
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/watcher"
	moduleplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano/pkg/log/vzlog"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	controllerruntime "sigs.k8s.io/controller-runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// ControllerConfig specifies the config of the controller using this base controller
type ControllerConfig struct {
	spi.Finalizer
	spi.Reconciler
	spi.Watcher
}

// Reconciler contains data needed to reconcile a DNS object.
type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Controller controller.Controller
	ControllerConfig
	LifecycleClass moduleplatform.LifecycleClassType

	// watcherMap is needed to keep track of which CRs have been initialized
	// key is the NSN of the Gateway
	watcherMap map[string][]watcher.WatchContext
}

// InitBaseController inits the base controller
func InitBaseController(mgr controllerruntime.Manager, controllerConfig ControllerConfig, class moduleplatform.LifecycleClassType) (*Reconciler, error) {
	r := Reconciler{
		Client:           mgr.GetClient(),
		Scheme:           mgr.GetScheme(),
		ControllerConfig: controllerConfig,
		watcherMap:       make(map[string][]watcher.WatchContext),
		LifecycleClass:   class,
	}

	var err error
	r.Controller, err = ctrl.NewControllerManagedBy(mgr).
		For(controllerConfig.Reconciler.GetReconcileObject()).
		WithEventFilter(r.createPredicateFilter()).
		Build(&r)

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
	mlc := moduleplatform.ModuleLifecycle{}
	objectkey := client.ObjectKeyFromObject(object)
	if err := r.Get(context.TODO(), objectkey, &mlc); err != nil {
		zap.S().Errorf("Failed to get ModuleLifecycle %s", objectkey)
		return false
	}
	return mlc.Spec.LifecycleClass == r.LifecycleClass
}
