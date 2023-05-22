// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package module

import (
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/handlerspi"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/statemachine"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/base/controllerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/util"
	"github.com/verrazzano/verrazzano-modules/pkg/semver"
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

	ctx := handlerspi.HandlerContext{Client: r.Client, Log: spictx.Log}
	handler, res := r.getActionHandler(ctx, cr)
	if res.Requeue {
		return res, nil
	}
	if handler == nil {
		return util.NewRequeueWithShortDelay(), nil
	}

	return r.reconcileAction(spictx, cr, handler)
}

// reconcileAction reconciles the Module CR for a particular action
func (r Reconciler) reconcileAction(spictx controllerspi.ReconcileContext, cr *moduleapi.Module, handler handlerspi.StateMachineHandler) (ctrl.Result, error) {
	ctx := handlerspi.HandlerContext{Client: r.Client, Log: spictx.Log}

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

// getActionHandler must return one of the Module action handlers.
func (r *Reconciler) getActionHandler(ctx handlerspi.HandlerContext, cr *moduleapi.Module) (handlerspi.StateMachineHandler, ctrl.Result) {
	if !hasInstalledCondition(ctx, cr) {
		return r.HandlerInfo.InstallActionHandler, ctrl.Result{}
	}

	// return UpgradeAction only when the desired version is different from current
	upgradeNeeded, err := IsUpgradeNeeded(cr.Spec.Version, cr.Status.Version)
	if err != nil {
		ctx.Log.ErrorfThrottled("Failed checking if upgrade needed for Module %s/%s failed with error: %v\n", cr.Namespace, cr.Name, err)
		return nil, util.NewRequeueWithShortDelay()
	}
	if upgradeNeeded {
		return r.HandlerInfo.UpgradeActionHandler, ctrl.Result{}
	}
	return r.HandlerInfo.UpdateActionHandler, ctrl.Result{}

}

func hasInstalledCondition(ctx handlerspi.HandlerContext, cr *moduleapi.Module) bool {
	for _, cond := range cr.Status.Conditions {
		if cond.Type == moduleapi.CondInstallComplete {
			return true
		}
	}
	return false
}

// IsUpgradeNeeded returns true if upgrade is needed
func IsUpgradeNeeded(desiredVersion, installedVersion string) (bool, error) {
	if len(desiredVersion) == 0 {
		return false, nil
	}
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
