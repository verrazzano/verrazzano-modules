// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package lifecycle

import (
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/spi"
	vzplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Reconcile updates the Certificate
func (r Reconciler) Reconcile(ctx spi.ReconcileContext, u *unstructured.Unstructured) (ctrl.Result, error) {
	cr := &vzplatform.ModuleLifecycle{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, cr); err != nil {
		return ctrl.Result{}, err
	}

	// TODO - process the lifecycle resource

	return ctrl.Result{}, nil
}
