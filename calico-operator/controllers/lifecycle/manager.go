// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package lifecycle

import (
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/baseconroller"
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/spi"
	corev1 "k8s.io/api/core/v1"
	ctrlruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Specify the SPI interfaces that this controller implements
var _ spi.DescribeController = Reconciler{}
var _ spi.ReconcileController = Reconciler{}
var _ spi.WatchController = Reconciler{}

type Reconciler struct {
	Client client.Client
}

var controller Reconciler

// InitController start the  controller
func InitController(mgr ctrlruntime.Manager) error {
	// The config MUST contain either a ReconcileController or a NoRulesController
	mcConfig := basecontroller.MicroControllerConfig{
		DescribeController:  &controller,
		ReconcileController: &controller,
		WatchController:     &controller,
	}
	br, err := basecontroller.InitBaseController(mgr, mcConfig)
	if err != nil {
		return err
	}
	controller.Client = br.Client
	return nil
}

// GetWatchedKinds returns the list of object kinds being watched
func (r Reconciler) GetWatchedKinds() []spi.WatchedKind {
	return []spi.WatchedKind{{
		Kind:                source.Kind{Type: &corev1.Service{}},
		FuncShouldReconcile: ShouldIngressEventTriggerReconcile,
	}}
}

// GetReconcileObject returns the kind of object being reconciled
func (r Reconciler) GetReconcileObject() client.Object {
	return &networkapi.DNS{}
}

// ShouldIngressEventTriggerReconcile returns true if a reconcile should be done
func ShouldIngressEventTriggerReconcile(obj client.Object, event spi.WatchEvent) bool {
	if event == spi.Deleted {
		return false
	}
	service := obj.(*corev1.Service)
	if service.Labels != nil {
		_, ok := service.Labels[constants.DnsIdLabel]
		return ok
	}
	return false
}
