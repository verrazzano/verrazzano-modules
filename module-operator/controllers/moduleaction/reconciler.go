// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package moduleaction

import (
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/handlerspi"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/statemachine"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/base/controllerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/util"
	"github.com/verrazzano/verrazzano-modules/pkg/k8s"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetReconcileObject returns the kind of object being reconciled
func (r Reconciler) GetReconcileObject() client.Object {
	return &moduleapi.ModuleAction{}
}

var executeStateMachine = defaultExecuteStateMachine

// Reconcile reconciles the ModuleAction ModuleCR
func (r Reconciler) Reconcile(spictx controllerspi.ReconcileContext, u *unstructured.Unstructured) (ctrl.Result, error) {
	cr := &moduleapi.ModuleAction{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, cr); err != nil {
		// This is a fatal internal error, don't requeue
		spictx.Log.ErrorfThrottled("Failed converting Unstructured to ModuleAction %s/%s: %v", err, u.GetNamespace(), u.GetName(), err)
		return util.NewRequeueWithShortDelay(), nil
	}
	nsn := k8s.GetNamespacedName(cr.ObjectMeta)

	// This is an imperative command, don't rerun it
	if cr.Status.State == moduleapi.StateCompleted || cr.Status.State == moduleapi.StateNotNeeded {
		spictx.Log.Oncef("Resource %v has already been processed, nothing to do", nsn)
		return ctrl.Result{}, nil
	}

	ctx := handlerspi.HandlerContext{Client: r.Client, Log: spictx.Log}

	helmInfo := loadHelmInfo(cr)
	handler, res := r.getActionHandler(ctx, cr)
	if res.Requeue {
		return res, nil
	}
	if handler == nil {
		return util.NewRequeueWithShortDelay(), nil
	}

	sm := statemachine.StateMachine{
		Scheme:   r.Scheme,
		CR:       cr,
		HelmInfo: &helmInfo,
		Handler:  handler,
	}

	res = executeStateMachine(ctx, sm)
	return res, nil
}

func loadHelmInfo(cr *moduleapi.ModuleAction) handlerspi.HelmInfo {
	helmInfo := handlerspi.HelmInfo{
		HelmRelease: cr.Spec.Installer.HelmRelease,
	}
	return helmInfo
}

// getActionHandler must return one of the MLC action handlers.
func (r *Reconciler) getActionHandler(ctx handlerspi.HandlerContext, cr *moduleapi.ModuleAction) (handlerspi.StateMachineHandler, ctrl.Result) {
	if cr.Spec.Action == moduleapi.DeleteAction {
		return r.handlerInfo.DeleteActionHandler, ctrl.Result{}
	}
	// Get the actual state of the module from the Kubernetes cluster
	state, res, err := r.handlerInfo.ModuleActualStateInCluster.GetActualModuleState(ctx, cr)
	if res2 := util.DeriveResult(res, err); res2.Requeue {
		return nil, res2
	}

	switch state {
	case handlerspi.ModuleStateNotInstalled:
		// install
		return r.handlerInfo.InstallActionHandler, ctrl.Result{}
	case handlerspi.ModuleStateReady:
		// the module is installed, if the version is changing this is an upgrade, else update
		upgrade, res, err := r.handlerInfo.ModuleActualStateInCluster.IsUpgradeNeeded(ctx, cr)
		if res2 := util.DeriveResult(res, err); res2.Requeue {
			return nil, res2
		}
		if upgrade {
			// upgrade
			return r.handlerInfo.UpgradeActionHandler, ctrl.Result{}
		}
		// update
		return r.handlerInfo.UpdateActionHandler, ctrl.Result{}
	default:
		ctx.Log.Progressf("Module is not is a state where any action can be taken %s/%s state: %s", state)
		return nil, ctrl.Result{}
	}
}

func defaultExecuteStateMachine(ctx handlerspi.HandlerContext, sm statemachine.StateMachine) ctrl.Result {
	return sm.Execute(ctx)
}
