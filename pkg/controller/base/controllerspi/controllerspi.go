// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package controllerspi

import (
	"context"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/result"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

// FuncShouldReconcile returns true if the watched object event should trigger reconcile
type FuncShouldReconcile func(reconciledResource types.NamespacedName, watchedObject client.Object, event WatchEventType) bool

// FuncControllerEventFilter is the predicate event handler filter that returns true if the object should be reconciled.
// This is needed to use same CR for multiple controllers
type FuncControllerEventFilter func(cli client.Client, object client.Object) bool

// WatchEventType is the type of watched event
type WatchEventType int

const (
	// Created indicates the watched object was created
	Created WatchEventType = iota

	// Updated indicates the watched object was updated
	Updated WatchEventType = iota

	// Deleted indicates the watched object was deleted
	Deleted WatchEventType = iota
)

// WatchDescriptor described an object being watched
type WatchDescriptor struct {
	// WatchedResourceKind is the kind of resource being watched
	WatchedResourceKind source.Kind

	// FuncShouldReconcile is called when watch event occurs to determine if CR should be reconciled
	FuncShouldReconcile
}

// WatchEvent is an occurrence of a watch event
type WatchEvent struct {
	// WatchEventType is the type of watched event
	WatchEventType

	// EventTime is the time the event occurred
	EventTime time.Time

	// WatchedResource is the resource that caused the event
	WatchedResource client.Object

	// ReconcilingResource is the resource that is potentially being reconciled
	ReconcilingResource types.NamespacedName
}

// ReconcileContext is a context has the dynamic context needed for a reconcile operation
type ReconcileContext struct {
	// Log is the VerrazzanoLogger
	Log vzlog.VerrazzanoLogger

	// ClientCtx is the context used to make controller runtime client API calls
	ClientCtx context.Context
}

// Reconciler is an interface used by controllers to reconcile a resource
type Reconciler interface {
	// Reconcile reconciles the resource
	Reconcile(ReconcileContext, *unstructured.Unstructured) result.Result

	// GetReconcileObject returns the client object being reconciled
	GetReconcileObject() client.Object
}

// EventFilter is an interface used by controllers filter events
type EventFilter interface {
	// HandlePredicateEvent is the predicate event handler filter that returns true if the object should be reconciled.
	// This is needed to use same CR for multiple controllers
	HandlePredicateEvent(cli client.Client, object client.Object) bool
}

// Watcher is an interface used by controllers that watch resources
type Watcher interface {
	// GetWatchDescriptors returns the list of WatchDescriptor for objects being watched
	GetWatchDescriptors() []WatchDescriptor
}

// Finalizer is an interface used by controllers the use finalizers
type Finalizer interface {
	// GetName returns the name of the finalizer
	GetName() string

	// PreRemoveFinalizer is called when the resource is being deleted, before the finalizer
	// is removed.  Use this method to delete Kubernetes resources, etc.
	PreRemoveFinalizer(ReconcileContext, *unstructured.Unstructured) result.Result

	// PostRemoveFinalizer is called after the finalizer is successfully removed.
	// This method does garbage collection and other tasks that can never return an error
	PostRemoveFinalizer(ReconcileContext, *unstructured.Unstructured)
}
