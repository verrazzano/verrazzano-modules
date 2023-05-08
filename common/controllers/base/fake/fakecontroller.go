// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package fake

import (
	"context"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	controllerruntime "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type FakeController struct{}

var _ controllerruntime.Controller = &FakeController{}

func (f FakeController) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	return ctrl.Result{}, nil
}

func (f FakeController) Watch(src source.Source, eventhandler handler.EventHandler, predicates ...predicate.Predicate) error {
	return nil
}

func (f FakeController) Start(ctx context.Context) error {
	return nil
}

func (f FakeController) GetLogger() logr.Logger {
	return logr.Logger{}
}
