// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package module

import (
	"github.com/verrazzano/verrazzano-modules/common/actionspi"
	"github.com/verrazzano/verrazzano-modules/common/controllercore/controllerspi"
	"github.com/verrazzano/verrazzano-modules/common/controllercore/statemachine"
	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	"github.com/verrazzano/verrazzano-modules/common/pkg/semver"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"strings"
	"time"

	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
)

// Reconcile reconciles the Module CR
func (r Reconciler) Reconcile(spictx controllerspi.ReconcileContext, u *unstructured.Unstructured) (ctrl.Result, error) {
	cr := &moduleapi.Module{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, cr); err != nil {
		return ctrl.Result{}, err
	}

	// TODO - there needs to be a check if a watch caused this to reconcile, the generation will be the same
	if cr.Status.ObservedGeneration == cr.Generation {
		return ctrl.Result{}, nil
	}

	handler, err := r.getActionHandler(cr)
	if err != nil {
		spictx.Log.ErrorfThrottled("Failed getting Lifecycle handler for Module %s/%s: %v", cr.Namespace, cr.Name, err)
		return util.NewRequeueWithShortDelay(), nil
	}
	if handler == nil {
		// Don't requeue, nothing to do
		return ctrl.Result{}, nil
	}
	return r.reconcileAction(spictx, cr, handler)
}

// reconcileAction reconciles the Module CR for a particular action
func (r Reconciler) reconcileAction(spictx controllerspi.ReconcileContext, cr *moduleapi.Module, handler actionspi.LifecycleActionHandler) (ctrl.Result, error) {
	ctx := actionspi.HandlerContext{Client: r.Client, Log: spictx.Log}

	helmInfo, err := loadHelmInfo(cr)
	if err != nil {
		if strings.Contains(err.Error(), "FileNotFound") {
			err := spictx.Log.ErrorfNewErr("Failed loading file information: %v", err)
			return util.NewRequeueWithDelay(10, 15, time.Second), err
		}
		err := spictx.Log.ErrorfNewErr("Failed loading Helm info for %s/%s: %v", cr.Namespace, cr.Name, err)
		return util.NewRequeueWithShortDelay(), err
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

func (r *Reconciler) getActionHandler(cr *moduleapi.Module) (actionspi.LifecycleActionHandler, error) {
	// Check for install complete
	if !isConditionPresent(cr, moduleapi.CondInstallComplete) {
		return r.comp.InstallActionHandler, nil
	}

	// return UpgradeAction only when the desired version is different from current
	upgradeNeeded, err := IsUpgradeNeeded(cr.Spec.Version, cr.Status.Version)
	if err != nil {
		return nil, err
	}
	if upgradeNeeded {
		return r.comp.UpgradeActionHandler, nil
	}

	// The module is already installed.  Check if update needed
	return r.comp.UpdateActionHandler, nil
}

func isConditionPresent(cr *moduleapi.Module, condition moduleapi.LifecycleCondition) bool {
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
