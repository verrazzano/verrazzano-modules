// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package controllerspi

import (
	"context"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrl "sigs.k8s.io/controller-runtime"
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

// WatchDescriptor described an object being watched
type WatchDescriptor struct {
	WatchKind source.Kind
	FuncShouldReconcile
}

// ReconcileContext is a context has the dynamic context needed for a reconcile operation
type ReconcileContext struct {
	Log       vzlog.VerrazzanoLogger
	ClientCtx context.Context
}

// Reconciler is an interface used by controllers to reconcile a resource
type Reconciler interface {
	// Reconcile reconciles the resource
	Reconcile(ReconcileContext, *unstructured.Unstructured) (ctrl.Result, error)

	// GetReconcileObject returns the client object being reconciled
	GetReconcileObject() client.Object
}

// Watcher is an interface used by controllers that watch resources
type Watcher interface {
	// GetWatchDescriptors returns the list of object kinds being watched
	GetWatchDescriptors() []WatchDescriptor
}

// Finalizer is an interface used by controllers the use finalizers
type Finalizer interface {
	// GetName returns the name of the finalizer
	GetName() string

	// PreRemoveFinalizer is called when the resource is being deleted, before the finalizer
	// is removed.  Use this method to delete Kubernetes resources, etc.
	PreRemoveFinalizer(ReconcileContext, *unstructured.Unstructured) (ctrl.Result, error)

	// PostRemoveFinalizer is called after the finalizer is successfully removed.
	// This method does garbage collection and other tasks that can never return an error
	PostRemoveFinalizer(ReconcileContext, *unstructured.Unstructured)
}
