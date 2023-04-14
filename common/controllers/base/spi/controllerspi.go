// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package spi

import (
	"context"
	"github.com/verrazzano/verrazzano/pkg/log/vzlog"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// FuncShouldReconcile returns true if the watched object event should trigger reconcile
type FuncShouldReconcile func(object client.Object, event WatchEvent) bool

type WatchEvent int

const (
	// Created indicates the watched object was created
	Created WatchEvent = iota

	// Updated indicates the watched object was updated
	Updated WatchEvent = iota

	// Deleted indicates the watched object was deleted
	Deleted WatchEvent = iota
)

// WatchedKind described an object being watched
type WatchedKind struct {
	Kind source.Kind
	FuncShouldReconcile
}

// ReconcileContext is a context has the dynamic context needed for a reconcile operation
type ReconcileContext struct {
	Log       vzlog.VerrazzanoLogger
	ClientCtx context.Context
}

// DescribeController is an interface used by controllers that only update a single resource kind
type DescribeController interface {
	// GetReconcileObject returns the client object being reconciled
	GetReconcileObject() client.Object
}

// ReconcileController is an interface used by controllers to reconcile a resource
type ReconcileController interface {
	// Reconcile reconciles the resource
	Reconcile(ReconcileContext, *unstructured.Unstructured) error
}

// WatchController is an interface used by controllers that watch resources
type WatchController interface {
	// GetWatchedKinds returns the list of object kinds being watched
	GetWatchedKinds() []WatchedKind
}

// FinalizerController is an interface used by controllers to manage finalizers
type FinalizerController interface {
	// AddFinalizer adds a finalizer
	AddFinalizer()

	// GarbageCollect garbage collects any related resources that were created by the controller
	GarbageCollect()

	// RemoveFinalizer removes a finalizer
	RemoveFinalizer()
}
