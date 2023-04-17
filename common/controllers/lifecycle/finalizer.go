// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package lifecycle

import (
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/spi"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrl "sigs.k8s.io/controller-runtime"
)

const finalizerName = "modulelifecycle.finalizer.verrazzano.io"

// GetName returns the name of the finalizer
func (r Reconciler) GetName() string {
	return finalizerName
}

// Cleanup garbage collects any related resources that were created by the controller
func (r Reconciler) Cleanup(spictx spi.ReconcileContext, u *unstructured.Unstructured) (ctrl.Result, error) {
	//if err := UpdateStatus(r.Client(), mlc, string(modulesv1alpha1.CondUninstall), modulesv1alpha1.CondUninstall); err != nil {
	//	return ctrl.Result{}, err
	//}
	//if err := r.Uninstall(ctx); err != nil {
	//	return newRequeueWithDelay(), err
	//}
	//if err := removeFinalizer(ctx, mlc); err != nil {
	//	return newRequeueWithDelay(), err
	//}
	//ctx.Log().Infof("Uninstall of %s complete", common.GetNamespacedName(mlc.ObjectMeta))
	//return ctrl.Result{}, nil

	return ctrl.Result{}, nil
}
