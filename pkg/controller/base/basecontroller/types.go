// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package basecontroller

import (
	"github.com/verrazzano/verrazzano-modules/pkg/controller/base/controllerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sync/atomic"
)

// ControllerConfig specifies the config of the controller using this base controller
type ControllerConfig struct {
	controllerspi.Finalizer
	controllerspi.Reconciler
	controllerspi.EventFilter
	controllerspi.Watcher
}

// Reconciler contains data needed to reconcile a DNS object.
type Reconciler struct {
	// Client is the controller-runtime client
	client.Client

	// Scheme is the CR scheme
	Scheme     *runtime.Scheme

	// Controller is a controller-runtime controller
	Controller controller.Controller

	// layeredControllerConfig config is the layered controller
	layeredControllerConfig ControllerConfig

	// WatchEventOccurred is a toggle to indicate that a watch has occurred. Once it is reconciled, then it is cleared.
	WatchEventOccurred atomic.Bool

	// watcherMap is used to determine if a watches have been initialized for the CR instance
	watcherMap         map[types.NamespacedName]bool

	// watchContexts is the list of watchContexts, one for each watch
	watchContexts      []*WatchContext
}

// WatchContext provides context to a watcher
// There is a WatchContext for each resource being watched by each instance of a CR.
type WatchContext struct {
	// Controller is a controller-runtime controller
	Controller              controller.Controller

	// Log is the Verrazzano logger
	Log                     vzlog.VerrazzanoLogger

	// watchDescriptor describes the resource being watched
	watchDescriptor 	controllerspi.WatchDescriptor

	// resourceBeingReconciled is the resource being reconciled
	resourceBeingReconciled types.NamespacedName
}
