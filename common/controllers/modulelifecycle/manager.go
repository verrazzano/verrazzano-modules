// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package modulelifecycle

import (
	"github.com/verrazzano/verrazzano-modules/common/actionspi"
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/basecontroller"
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/spi"
	moduleplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrlruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconciler contains the fields needed by the layered controller.
type Reconciler struct {
	// Client is the controller runtime client.
	Client client.Client

	// Scheme is the scheme for this controller.  I must contain the schemes of all Kubernetes API objects that it accesses
	// through the client.
	Scheme *runtime.Scheme

	// handlers contains the action handlers
	handlers actionspi.ActionHandlers
}

// Specify the SPI interfaces that this controller implements
var _ spi.Reconciler = Reconciler{}
var controller Reconciler

// InitController start the  controller
func InitController(mgr ctrlruntime.Manager, comp actionspi.ActionHandlers, class moduleplatform.LifecycleClassType) error {
	// The config MUST contain at least the Reconciler.  Other spi interfaces are optional.
	config := basecontroller.ControllerConfig{
		Reconciler: &controller,
		Finalizer:  &controller,
	}
	baseController, err := basecontroller.InitBaseController(mgr, config, class)
	if err != nil {
		return err
	}

	// init other controller fields
	controller.Client = baseController.Client
	controller.Scheme = baseController.Scheme
	controller.handlers = comp
	return nil
}
