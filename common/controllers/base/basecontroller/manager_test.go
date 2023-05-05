// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package basecontroller

import (
	"github.com/stretchr/testify/assert"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano/pkg/k8sutil"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"testing"
)

// TestInitBaseController tests the initialization of the base controller
// GIVEN a base  controller
// WHEN InitBaseController is called
// THEN the base controller should be properly initialized
func TestInitBaseController(t *testing.T) {
	asserts := assert.New(t)

	mgr, err := controllerruntime.NewManager(k8sutil.GetConfigOrDieFromController(), controllerruntime.Options{
		Scheme: newScheme(),
		Port:   8080,
	})
	asserts.NoError(err)
	config := ControllerConfig{
		Reconciler: &ReconcilerImpl{},
		Finalizer:  &FinalizerImpl{},
	}

	r, err := InitBaseController(mgr, config, moduleapi.CalicoLifecycleClass)
	asserts.NoError(err)
	asserts.NotNil(r)
}
