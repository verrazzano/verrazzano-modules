// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package module

import (
	compspi "github.com/verrazzano/verrazzano-modules/common/actionspi"
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/spi"
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/statemachine"
	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	"github.com/verrazzano/verrazzano/tests/e2e/pkg"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"strings"
	"time"

	moduleplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"

	vzspi "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
)

// Reconcile reconciles the Module CR
func (r Reconciler) Reconcile(spictx spi.ReconcileContext, u *unstructured.Unstructured) (ctrl.Result, error) {
	cr := &moduleplatform.Module{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, cr); err != nil {
		return ctrl.Result{}, err
	}
	handler := r.getActionHandler(cr)
	if handler == nil {
		spictx.Log.Errorf("Not a valid Action")
		// Dont requeue, this is a fatal error
		return ctrl.Result{}, nil
	}
	return r.reconcileAction(spictx, cr, handler)
}

// reconcileAction reconciles the Module CR for a particular action
func (r Reconciler) reconcileAction(spictx spi.ReconcileContext, cr *moduleplatform.Module, handler compspi.LifecycleActionHandler) (ctrl.Result, error) {
	ctx, err := vzspi.NewMinimalContext(r.Client, spictx.Log)
	if err != nil {
		return util.NewRequeueWithShortDelay(), err
	}

	helmInfo, err := loadHelmInfo(cr)
	if err != nil {
		if strings.Contains(err.Error(), "FileNotFound") {
			spictx.Log.ErrorfNewErr("Failed loading file information: %v", err)
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

func (r *Reconciler) getActionHandler(cr *moduleplatform.Module) compspi.LifecycleActionHandler {
	// Check for install complete
	if !isConditionPresent(cr, moduleplatform.CondInstallComplete) {
		return r.comp.InstallActionHandler
	}
	// return UpgradeAction only when the desired version is different from current
	isGreaterVersion, err := pkg.IsMinVersion(cr.Spec.Version, cr.Status.Version)
	if err != nil {
		return nil
	}
	if isGreaterVersion {
		return r.comp.UpgradeActionHandler
	}
	return r.comp.InstallActionHandler
}

func isConditionPresent(cr *moduleplatform.Module, condition moduleplatform.LifecycleCondition) bool {
	for _, each := range cr.Status.Conditions {
		if each.Type == condition {
			return true
		}
	}
	return false
}
