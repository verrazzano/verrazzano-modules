// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package module

import (
	"context"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/basecontroller"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/spi/controllerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/spi/handlerspi"
	ctrlruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// Specify the SPI interfaces that this controller implements
var _ controllerspi.Reconciler = Reconciler{}

type Reconciler struct {
	*basecontroller.BaseReconciler
	ModuleControllerConfig
}

type ModuleControllerConfig struct {
	ControllerManager ctrlruntime.Manager
	ControllerOptions *controller.Options
	ModuleHandlerInfo handlerspi.ModuleHandlerInfo
	ModuleClass       moduleapi.ModuleClassType
	WatchDescriptors  []controllerspi.WatchDescriptor
}

// InitController start the  controller
func InitController(moduleConfig ModuleControllerConfig) error {
	moduleController := Reconciler{}

	opt := controller.Options{}
	if moduleConfig.ControllerOptions != nil {
		opt = *moduleConfig.ControllerOptions
	}
	// The config MUST contain at least the BaseReconciler.  Other spi interfaces are optional.
	config := basecontroller.ControllerConfig{
		Reconciler:  &moduleController,
		Finalizer:   &moduleController,
		EventFilter: &moduleController,
		Watcher:     &moduleController,
		Options:     opt,
	}

	baseReconciler, err := basecontroller.CreateControllerAndAddItToManager(moduleConfig.ControllerManager, config)
	if err != nil {
		return err
	}
	moduleController.BaseReconciler = baseReconciler

	// init other controller fields
	moduleController.ModuleControllerConfig = moduleConfig
	return nil
}

// GetReconcileObject returns the kind of object being reconciled
func (r Reconciler) GetReconcileObject() client.Object {
	return &moduleapi.Module{}
}

func (r Reconciler) HandlePredicateEvent(cli client.Client, object client.Object) bool {
	mlc := moduleapi.Module{}
	objectkey := client.ObjectKeyFromObject(object)
	if err := cli.Get(context.TODO(), objectkey, &mlc); err != nil {
		return false
	}
	return mlc.Spec.ModuleName == string(r.ModuleClass)
}

// GetWatchDescriptors returns the user supplied watch descriptors along with the default Module ones
func (r Reconciler) GetWatchDescriptors() []controllerspi.WatchDescriptor {
	return append(r.GetDefaultWatchDescriptors(), r.WatchDescriptors...)
}
