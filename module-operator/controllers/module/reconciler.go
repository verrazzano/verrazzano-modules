// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package module

import (
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/spi"
	compspi "github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/action_spi"
	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

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
		err := spictx.Log.ErrorfNewErr("Failed loading Helm info for %s/%s: %v", cr.Namespace, cr.Name, err)
		return util.NewRequeueWithShortDelay(), err
	}
	tracker := getTracker(cr.ObjectMeta, stateInit)

	smc := stateMachineContext{
		cr:        cr,
		tracker:   tracker,
		chartInfo: &helmInfo,
		handler:   handler,
	}

	res := r.doStateMachine(ctx, smc)
	return res, nil
}

func (r *Reconciler) getActionHandler(cr *moduleplatform.Module) compspi.LifecycleActionHandler {

	return r.comp.InstallAction
}
