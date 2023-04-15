// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package lifecycle

import (
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/spi"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Reconcile updates the Certificate
func (r Reconciler) Reconcile(ctx spi.ReconcileContext, u *unstructured.Unstructured) error {
	//dns := &networkapi.DNS{}
	//if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, dns); err != nil {
	//	return err
	//}

	return nil
}
