// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package modulelifecycle

import (
	"github.com/verrazzano/verrazzano-modules/common/controllercore/controllerspi"
	"github.com/verrazzano/verrazzano-modules/common/controllercore/statemachine"
	"github.com/verrazzano/verrazzano-modules/common/handlerspi"
	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	"github.com/verrazzano/verrazzano-modules/common/pkg/k8s"
	"github.com/verrazzano/verrazzano-modules/common/pkg/semver"
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
		spictx.Log.ErrorfThrottled("Failed converting Unstructured to ModuleLifecycle %s/%s: %v", err, u.GetNamespace(), u.GetName(), err)
		return util.NewRequeueWithShortDelay(), nil
	}
	nsn := k8s.GetNamespacedName(cr.ObjectMeta)

	// This is an imperative command, don't rerun it
	if cr.Status.State == moduleapi.StateCompleted || cr.Status.State == moduleapi.StateNotNeeded {
		spictx.Log.Oncef("Resource %v has already been processed, nothing to do", nsn)
		return ctrl.Result{}, nil
	}

	helmInfo := loadHelmInfo(cr)
	handler, err := r.getActionHandler(cr)
	if err != nil {
		spictx.Log.ErrorfThrottled("Failed checking ModuleLifecycle CR %/%s version: %v", err, u.GetNamespace(), u.GetName(), err)
		return util.NewRequeueWithShortDelay(), nil
	}

	sm := statemachine.StateMachine{
		Scheme:   r.Scheme,
		CR:       cr,
		HelmInfo: &helmInfo,
		Handler:  handler,
	}
	ctx := handlerspi.HandlerContext{Client: r.Client, Log: spictx.Log}

	res := executeStateMachine(sm, ctx)
	return res, nil
}

func loadHelmInfo(cr *moduleapi.ModuleLifecycle) handlerspi.HelmInfo {
	helmInfo := handlerspi.HelmInfo{
		HelmRelease: cr.Spec.Installer.HelmRelease,
	}
	return helmInfo
}

func (r *Reconciler) getActionHandler(cr *moduleapi.ModuleLifecycle) (handlerspi.StateMachineHandler, error) {
	// Check for install complete
	if !isConditionPresent(cr, moduleapi.CondInstallComplete) {
		return r.handlerInfo.InstallActionHandler, nil
	}

	// return UpgradeAction only when the desired version is different from current
	upgradeNeeded, err := IsUpgradeNeeded(cr.Spec.Version, cr.Status.Version)
	if err != nil {
		return nil, err
	}

	if upgradeNeeded {
		return r.handlerInfo.UpgradeActionHandler, nil
	}

	// The module is already installed.  Check if update needed
	return r.handlerInfo.UpdateActionHandler, nil
}

func defaultExecuteStateMachine(sm statemachine.StateMachine, ctx handlerspi.HandlerContext) ctrl.Result {
	return sm.Execute(ctx)
}

func isConditionPresent(cr *moduleapi.ModuleLifecycle, condition moduleapi.LifecycleCondition) bool {
	for _, each := range cr.Status.Conditions {
		if each.Type == condition {
			return true
		}
	}
	return false
}

// IsUpgradeNeeded returns true if upgrade is needed
func IsUpgradeNeeded(desiredVersion, installedVersion string) (bool, error) {
	desiredSemver, err := semver.NewSemVersion(desiredVersion)
	if err != nil {
		return false, err
	}
	installedSemver, err := semver.NewSemVersion(installedVersion)
	if err != nil {
		return false, err
	}
	return installedSemver.IsLessThan(desiredSemver), nil
}
