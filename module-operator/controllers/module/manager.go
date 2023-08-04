// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package module

import (
	"context"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/base/basecontroller"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/base/controllerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/handlerspi"
	ctrlruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Specify the SPI interfaces that this controller implements
var _ controllerspi.Reconciler = Reconciler{}

type Reconciler struct {
	*basecontroller.BaseReconciler
	ModuleControllerConfig
}

type ModuleControllerConfig struct {
	ControllerManager ctrlruntime.Manager
	ModuleHandlerInfo handlerspi.ModuleHandlerInfo
	ModuleClass       moduleapi.ModuleClassType
	WatchDescriptors  []controllerspi.WatchDescriptor
}

// InitController start the  controller
func InitController(modConfig ModuleControllerConfig) error {
	controller := Reconciler{}

	// The config MUST contain at least the BaseReconciler.  Other spi interfaces are optional.
	config := basecontroller.ControllerConfig{
		Reconciler:  &controller,
		Finalizer:   &controller,
		EventFilter: &controller,
		Watcher:     &controller,
	}

	baseReconciler, err := basecontroller.CreateControllerAndAddItToManager(modConfig.ControllerManager, config)
	if err != nil {
		return err
	}
	controller.BaseReconciler = baseReconciler

	// init other controller fields
	controller.ModuleControllerConfig = modConfig
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
