// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package modulelifecycle

import (
	"github.com/verrazzano/verrazzano-modules/common/actionspi"
	"github.com/verrazzano/verrazzano-modules/common/controllercore/controllerspi"
	"github.com/verrazzano/verrazzano-modules/common/controllercore/statemachine"
	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	"github.com/verrazzano/verrazzano-modules/common/pkg/k8s"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetReconcileObject returns the kind of object being reconciled
func (r Reconciler) GetReconcileObject() client.Object {
	return &moduleapi.ModuleLifecycle{}
}

var executeStateMachine = defaultExecuteStateMachine

// Reconcile reconciles the ModuleLifecycle ModuleCR
func (r Reconciler) Reconcile(spictx controllerspi.ReconcileContext, u *unstructured.Unstructured) (ctrl.Result, error) {
	cr := &moduleapi.ModuleLifecycle{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, cr); err != nil {
		// This is a fatal internal error, don't requeue
		spictx.Log.ErrorfThrottled("Failed converting Unstructured to ModuleLifecycle %s/%s: %v", err, u.GetNamespace(), u.GetName())
		return util.NewRequeueWithShortDelay(), nil
	}
	nsn := k8s.GetNamespacedName(cr.ObjectMeta)

	// This is an imperative command, don't rerun it
	if cr.Status.State == moduleapi.StateCompleted || cr.Status.State == moduleapi.StateNotNeeded {
		spictx.Log.Oncef("Resource %v has already been processed, nothing to do", nsn)
		return ctrl.Result{}, nil
	}

	helmInfo := loadHelmInfo(cr)
	handler := r.getActionHandler(cr.Spec.Action)
	if handler == nil {
		spictx.Log.ErrorfThrottled("Failed, internal error invalid ModuleLifecycle ModuleCR handler %s", cr.Spec.Action)
		return util.NewRequeueWithShortDelay(), nil
	}

	sm := statemachine.StateMachine{
		Scheme:   r.Scheme,
		CR:       cr,
		HelmInfo: &helmInfo,
		Handler:  handler,
	}
	ctx := actionspi.HandlerContext{Client: r.Client, Log: spictx.Log}

	res := executeStateMachine(sm, ctx)
	return res, nil
}

func loadHelmInfo(cr *moduleapi.ModuleLifecycle) actionspi.HelmInfo {
	helmInfo := actionspi.HelmInfo{
		HelmRelease: cr.Spec.Installer.HelmRelease,
	}
	return helmInfo
}

func (r *Reconciler) getActionHandler(action moduleapi.ActionType) actionspi.LifecycleActionHandler {
	switch action {
	case moduleapi.InstallAction:
		return r.handlers.InstallActionHandler
	case moduleapi.UninstallAction:
		return r.handlers.UninstallActionHandler
	case moduleapi.UpdateAction:
		return r.handlers.UpdateActionHandler
	case moduleapi.UpgradeAction:
		return r.handlers.UpgradeActionHandler
	default:
		return nil
	}
}

func defaultExecuteStateMachine(sm statemachine.StateMachine, ctx actionspi.HandlerContext) ctrl.Result {
	return sm.Execute(ctx)
}
