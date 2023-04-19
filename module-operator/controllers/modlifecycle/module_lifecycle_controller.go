// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package modlifecycle

import (
	"context"
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"

	modulesv1alpha1 "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/common"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/modlifecycle/delegates"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/modlifecycle/reconciler"
	controller2 "github.com/verrazzano/verrazzano-modules/module-operator/pkg/controller"
	"github.com/verrazzano/verrazzano/pkg/log/vzlog"
	"go.uber.org/zap"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Controller controller.Controller
}

const POCLifecycleClass = "helmpoc"

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&modulesv1alpha1.ModuleLifecycle{}).
		WithEventFilter(r.createPredicateFilter()).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 10,
		}).
		Complete(r)
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// Get the mlc for the request
	mlc := &modulesv1alpha1.ModuleLifecycle{}
	if err := r.Get(ctx, req.NamespacedName, mlc); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		zap.S().Errorf("Failed to get Module %s/%s", req.Namespace, req.Name)
		return newRequeueWithDelay(), err
	}
	// Get the resource logger needed to log message using 'progress' and 'once' methods
	log, err := vzlog.EnsureResourceLogger(&vzlog.ResourceConfig{
		Name:           mlc.Name,
		Namespace:      mlc.Namespace,
		ID:             string(mlc.UID),
		Generation:     mlc.Generation,
		ControllerName: "mlc-lifecycle",
	})
	if err != nil {
		zap.S().Errorf("Failed to create controller logger for ModuleLifecycle controller: %v", err)
		return newRequeueWithDelay(), err
	}

	// NOTE: Need to see if these be broken out into separate lifecycle operators
	delegate, err := reconciler.New(mlc, r.Status())
	if err != nil {
		// Unknown mlc controller cannot be handled; no need to re-reconcile until the resource is updated
		msg := fmt.Sprintf("Error retrieving delegate lifecycle reconciler: %v", err.Error())
		if err := reconciler.UpdateStatus(r.Client, mlc, msg, modulesv1alpha1.CondFailed); err != nil {
			return newRequeueWithDelay(), err
		}
		return ctrl.Result{}, err
	}

	if delegate == nil {
		return ctrl.Result{}, fmt.Errorf("No delegate found for mlc %s/%s", mlc.Namespace, mlc.Name)
	}
	if err != nil {
		return controller2.NewRequeueWithDelay(2, 5, time.Second), err
	}

	result, err := delegate.Reconcile(log, r.Client, mlc)
	if err != nil {
		return handleError(log, mlc, err)
	}
	return result, nil
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
	mlc := modulesv1alpha1.ModuleLifecycle{}
	objectkey := client.ObjectKeyFromObject(object)
	if err := r.Get(context.TODO(), objectkey, &mlc); err != nil {
		zap.S().Errorf("Failed to get ModuleLifecycle %s", objectkey)
		return false
	}
	handlesEvent := mlc.Spec.LifecycleClass == POCLifecycleClass
	zap.S().Debugf("POC Helm controller event filter result for %s: %v", objectkey, handlesEvent)
	return handlesEvent
}

func handleError(log vzlog.VerrazzanoLogger, mlc *modulesv1alpha1.ModuleLifecycle, err error) (ctrl.Result, error) {
	if k8serrors.IsConflict(err) {
		log.Debugf("Conflict resolving module lifecycle %s", common.GetNamespacedName(mlc.ObjectMeta))
	} else if delegates.IsNotReadyError(err) {
		log.Progressf("Module lifecycle %s is not ready yet", common.GetNamespacedName(mlc.ObjectMeta))
	} else {
		log.Errorf("failed to reconcile module %s: %v", common.GetNamespacedName(mlc.ObjectMeta), err)
		return newRequeueWithDelay(), err
	}
	return newRequeueWithDelay(), nil
}

func newRequeueWithDelay() reconcile.Result {
	return controller2.NewRequeueWithDelay(2, 3, time.Second)
}
