// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package moduleaction

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/base/controllerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	"testing"
)

// TestFinalizer tests that the finalizer implementation works correctly
// GIVEN a Finalizer
// WHEN the Finalizer methods are called
// THEN ensure that they work correctly
func TestFinalizer(t *testing.T) {
	asserts := assert.New(t)

	rctx := controllerspi.ReconcileContext{
		Log:       vzlog.DefaultLogger(),
		ClientCtx: context.TODO(),
	}
	r := Reconciler{}
	asserts.Equal(finalizerName, r.GetName())

	res, err := r.PreRemoveFinalizer(rctx, nil)
	asserts.NoError(err)
	asserts.False(res.Requeue)

	res, err = r.PreRemoveFinalizer(rctx, nil)
	asserts.NoError(err)
	asserts.False(res.Requeue)
}