// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package module

import (
	"github.com/verrazzano/verrazzano-modules/common/actionspi"
	"github.com/verrazzano/verrazzano-modules/common/controllercore/basecontroller"
	spi "github.com/verrazzano/verrazzano-modules/common/controllercore/controllerspi"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrlruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Specify the SPI interfaces that this controller implements
var _ spi.Reconciler = Reconciler{}

type Reconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
	comp   actionspi.ActionHandlers
}

var _ spi.Reconciler = Reconciler{}

var controller Reconciler

// InitController start the  controller
func InitController(mgr ctrlruntime.Manager, comp actionspi.ActionHandlers, class moduleapi.LifecycleClassType) error {
	// The config MUST contain at least the Reconciler.  Other spi interfaces are optional.
	config := basecontroller.ControllerConfig{
		Reconciler: &controller,
		Finalizer:  &controller,
	}
	baseController, err := basecontroller.CreateControllerAndAddItToManager(mgr, config, class)
	if err != nil {
		return err
	}

	// init other controller fields
	controller.Client = baseController.Client
	controller.Scheme = baseController.Scheme
	controller.comp = comp
	return nil
}

// GetReconcileObject returns the kind of object being reconciled
func (r Reconciler) GetReconcileObject() client.Object {
	return &moduleapi.Module{}
}
