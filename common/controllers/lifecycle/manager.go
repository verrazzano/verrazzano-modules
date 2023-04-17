// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package lifecycle

import (
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/basecontroller"
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/spi"
	vzplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	vzspi "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	ctrlruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Specify the SPI interfaces that this controller implements
var _ spi.Reconciler = Reconciler{}

type Reconciler struct {
	Client client.Client
	comp vzspi.Component
}

var _ spi.Reconciler = Reconciler{}


var controller Reconciler

// InitController start the  controller
func InitController(mgr ctrlruntime.Manager) error {
	// The config MUST contain at least a Reconciler.  Other spi interfaces are optional.
	config := basecontroller.ControllerConfig{
		Reconciler: &controller,
	}
	br, err := basecontroller.InitBaseController(mgr, config)
	if err != nil {
		return err
	}
	controller.Client = br.Client
	return nil
}

// GetReconcileObject returns the kind of object being reconciled
func (r Reconciler) GetReconcileObject() client.Object {
	return &vzplatform.ModuleLifecycle{}
}
