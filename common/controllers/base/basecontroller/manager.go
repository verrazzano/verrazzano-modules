// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package basecontroller

import (
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/spi"
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/watcher"
	"github.com/verrazzano/verrazzano/pkg/log/vzlog"
	"k8s.io/apimachinery/pkg/runtime"
	controllerruntime "sigs.k8s.io/controller-runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// ControllerConfig specifies the config of the controller using this base controller
type ControllerConfig struct {
	spi.Finalizer
	spi.Reconciler
	spi.Watcher
}

// Reconciler contains data needed to reconcile a DNS object.
type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Controller controller.Controller
	ControllerConfig

	// watcherMap is needed to keep track of which CRs have been initialized
	// key is the NSN of the Gateway
	watcherMap map[string][]watcher.WatchContext
}

// InitBaseController inits the base controller
func InitBaseController(mgr controllerruntime.Manager, controllerConfig ControllerConfig) (*Reconciler, error) {
	r := Reconciler{
		Client:           mgr.GetClient(),
		Scheme:           mgr.GetScheme(),
		ControllerConfig: controllerConfig,
		watcherMap:       make(map[string][]watcher.WatchContext),
	}

	var err error
	r.Controller, err = ctrl.NewControllerManagedBy(mgr).
		For(controllerConfig.Reconciler.GetReconcileObject()).Build(&r)

	if err != nil {
		return nil, vzlog.DefaultLogger().ErrorfNewErr("Failed calling SetupWithManager for Istio Gateway controller: %v", err)
	}
	return &r, nil
}
