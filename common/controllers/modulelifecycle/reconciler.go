// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package modulelifecycle

import (
	"github.com/verrazzano/verrazzano-modules/common/actionspi"
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/spi"
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/statemachine"
	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	"github.com/verrazzano/verrazzano-modules/common/pkg/k8s"
	moduleplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	vzspi "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetReconcileObject returns the kind of object being reconciled
func (r Reconciler) GetReconcileObject() client.Object {
	return &moduleplatform.ModuleLifecycle{}
}

// Reconcile reconciles the ModuleLifecycle CR
func (r Reconciler) Reconcile(spictx spi.ReconcileContext, u *unstructured.Unstructured) (ctrl.Result, error) {
	cr := &moduleplatform.ModuleLifecycle{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, cr); err != nil {
		return ctrl.Result{}, err
	}
	nsn := k8s.GetNamespacedName(cr.ObjectMeta)

	// This is an imperative command, don't rerun it
	if cr.Status.State == moduleplatform.StateCompleted || cr.Status.State == moduleplatform.StateNotNeeded {
		spictx.Log.Oncef("Resource %v has already been processed, nothing to do", nsn)
		return ctrl.Result{}, nil
	}

	ctx, err := vzspi.NewMinimalContext(r.Client, spictx.Log)
	if err != nil {
		return util.NewRequeueWithShortDelay(), err
	}

	if cr.Generation == cr.Status.ObservedGeneration {
		spictx.Log.Debugf("Skipping reconcile for %v, observed generation has not change", nsn)
		return util.NewRequeueWithShortDelay(), err
	}

	helmInfo := loadHelmInfo(cr)
	handler := r.getActionHandler(cr.Spec.Action)
	if handler == nil {
		spictx.Log.Errorf("Invalid ModuleLifecycle CR handler %s", cr.Spec.Action)
		// Dont requeue, this is a fatal error
		return ctrl.Result{}, nil
	}

	sm := statemachine.StateMachine{
		Scheme:   r.Scheme,
		CR:       cr,
		HelmInfo: &helmInfo,
		Handler:  handler,
	}

	res := sm.Execute(ctx)
	return res, nil
}

func loadHelmInfo(cr *moduleplatform.ModuleLifecycle) actionspi.HelmInfo {
	helmInfo := actionspi.HelmInfo{
		HelmRelease: cr.Spec.Installer.HelmRelease,
	}
	return helmInfo
}

func (r *Reconciler) getActionHandler(action moduleplatform.ActionType) actionspi.LifecycleActionHandler {
	switch action {
	case moduleplatform.InstallAction:
		return r.comp.InstallAction
	case moduleplatform.UninstallAction:
		return r.comp.UninstallAction
	case moduleplatform.UpdateAction:
		return r.comp.UpdateAction
	case moduleplatform.UpgradeAction:
		return r.comp.UpgradeAction
	}
	return nil
}
