// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package modulelifecycle

import (
	"context"
	"fmt"

	"github.com/verrazzano/verrazzano-modules/common/actionspi"
	"github.com/verrazzano/verrazzano-modules/common/controllercore/controllerspi"
	"github.com/verrazzano/verrazzano-modules/common/controllercore/statemachine"
	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	"github.com/verrazzano/verrazzano-modules/common/pkg/k8s"
	"github.com/verrazzano/verrazzano-modules/common/pkg/semver"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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

	handler, err := r.getActionHandler(cr, spictx)
	if err != nil {
		spictx.Log.ErrorfThrottled("Failed, internal error getting ModuleLifecycle ModuleCR handler %s", cr.Spec.Action)
		return util.NewRequeueWithShortDelay(), nil
	}

	if handler == nil {
		spictx.Log.ErrorfThrottled("Failed, internal error invalid ModuleLifecycle ModuleCR handler %s", cr.Spec.Action)
		return util.NewRequeueWithShortDelay(), nil
	}

	helmInfo := loadHelmInfo(cr)
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

/*func (r *Reconciler) getActionHandler(action moduleapi.ActionType) actionspi.LifecycleActionHandler {
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
}*/

func defaultExecuteStateMachine(sm statemachine.StateMachine, ctx actionspi.HandlerContext) ctrl.Result {
	return sm.Execute(ctx)
}

func (r *Reconciler) getActionHandler(cr *moduleapi.ModuleLifecycle, spictx controllerspi.ReconcileContext) (actionspi.LifecycleActionHandler, error) {
	moduleReferences := cr.GetOwnerReferences()
	if len(moduleReferences) == 0 {
		return nil, fmt.Errorf("no modules associated with modulelifecycle %s/%s", cr.GetNamespace(), cr.GetName())
	}

	if len(moduleReferences) > 1 {
		var moduleNames []string
		for _, moduleRef := range moduleReferences {
			moduleNames = append(moduleNames, moduleRef.Name)
		}
		return nil, fmt.Errorf("invalid modulelifecycle %s/%s with multiple modules %v", cr.GetNamespace(), cr.GetName(), moduleNames)
	}

	module, err := getModule(moduleReferences[0].Name, cr.GetNamespace(), r.Client, spictx)
	if err != nil {
		return nil, fmt.Errorf("error getting module %s: %w", moduleReferences[0].Name, err)
	}

	// Check for delete requested
	if module.GetDeletionTimestamp() != nil || cr.Spec.Action == moduleapi.ModuleLifecycleUninstallAction {
		return r.handlers.UninstallActionHandler, nil
	}

	if !isConditionPresent(module, moduleapi.CondInstallComplete) {
		return r.handlers.InstallActionHandler, nil
	}

	// return UpgradeAction only when the desired version is different from current
	upgradeNeeded, err := IsUpgradeNeeded(module.Spec.Version, module.Status.Version)
	if err != nil {
		return nil, err
	}

	if upgradeNeeded {
		return r.handlers.UpgradeActionHandler, nil
	}

	// The module is already installed.  Check if update needed
	return r.handlers.UpdateActionHandler, nil
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

func getModule(name string, namespace string, c client.Client, spictx controllerspi.ReconcileContext) (*moduleapi.Module, error) {
	m := moduleapi.Module{}
	nsn := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	if err := c.Get(context.TODO(), nsn, &m); err != nil {
		spictx.Log.Progressf("Retrying get for Module %v: %v", nsn, err)
		return nil, err
	}
	return &m, nil
}
