// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package lifecycle

import (
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/basecontroller"
	spi "github.com/verrazzano/verrazzano-modules/common/controllers/base/spi"
	compspi "github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/action_spi"
	moduleplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"

	ctrlruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Specify the SPI interfaces that this controller implements
var _ spi.Reconciler = Reconciler{}

type Reconciler struct {
	Client client.Client
	comp   compspi.LifecycleComponent
}

var _ spi.Reconciler = Reconciler{}

var controller Reconciler

// InitController start the  controller
func InitController(mgr ctrlruntime.Manager, comp compspi.LifecycleComponent, class moduleplatform.LifecycleClassType) error {
	// The config MUST contain at least the Reconciler.  Other spi interfaces are optional.
	config := basecontroller.ControllerConfig{
		Reconciler: &controller,
		Finalizer:  &controller,
	}
	br, err := basecontroller.InitBaseController(mgr, config, class)
	if err != nil {
		return err
	}

	// init other controller fields
	controller.Client = br.Client
	controller.comp = comp
	return nil
}

// GetReconcileObject returns the kind of object being reconciled
func (r Reconciler) GetReconcileObject() client.Object {
	return &moduleplatform.ModuleLifecycle{}
}
