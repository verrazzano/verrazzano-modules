// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package moduleaction

import (
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/handlerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/base/basecontroller"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/base/controllerspi"
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

	// handlerInfo contains the ModuleAction handler information
	handlerInfo handlerspi.ModuleLifecycleHandlerInfo

	// ModuleClassName is the class name of the controller
	ClassName moduleapi.ModuleClassType
}

// Specify the SPI interfaces that this controller implements
var _ controllerspi.Reconciler = Reconciler{}

// InitController start the  controller
func InitController(mgr ctrlruntime.Manager, handlers handlerspi.ModuleLifecycleHandlerInfo, className moduleapi.ModuleClassType) error {
	var controller Reconciler

	// The config MUST contain at least the Reconciler.  Other spi interfaces are optional.
	config := basecontroller.ControllerConfig{
		Reconciler: &controller,
		Finalizer:  &controller,
	}
	baseController, err := basecontroller.CreateControllerAndAddItToManager(mgr, config, className)
	if err != nil {
		return err
	}

	// init other controller fields
	controller.ClassName = className
	controller.Client = baseController.Client
	controller.Scheme = baseController.Scheme
	controller.handlerInfo = handlers
	return nil
}
