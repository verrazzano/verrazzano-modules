// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package statemachine

import (
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

// stateTracker keeps an in-memory state for a component doing actions
type stateTracker struct {
	state state
	gen   int64
}

// trackerContext contains context used to track the state of multiple resources
type trackerContext struct {
	// trackerMap has a map of trackers with key from VZ name, namespace, and UID
	trackerMap map[string]*stateTracker

	// trackerMutex is used to access the map concurrently
	trackerMutex sync.RWMutex
}

// newTrackerContext creates a new trackerContext
func newTrackerContext() *trackerContext {
	return &trackerContext{
		trackerMap:   make(map[string]*stateTracker),
		trackerMutex: sync.RWMutex{},
	}
}

// getTrackerKey gets the stateTracker key for the Verrazzano resource
func (t *trackerContext) getTrackerKey(CR client.Object, gen int64) string {
	return fmt.Sprintf("%s-%s-%v-%s", CR.GetNamespace(), CR.GetName(), gen, string(CR.GetUID()))
}

// ensureTracker gets the stateTracker, creating a new one if needed
func (t *trackerContext) ensureTracker(CR client.Object, initialState state) *stateTracker {
	t.trackerMutex.Lock()
	defer t.trackerMutex.Unlock()
	key := t.getTrackerKey(CR, CR.GetGeneration())
	tracker, ok := t.trackerMap[key]
	if ok {
		return tracker
	}

	// create a new tracker and save it in the map
	tracker = &stateTracker{
		state: initialState,
		gen:   CR.GetGeneration(),
	}
	t.trackerMap[key] = tracker

	// Delete the previous entry if it exists
	if CR.GetGeneration() == 1 {
		return tracker
	}
	keyPrev := t.getTrackerKey(CR, CR.GetGeneration()-1)
	_, ok = t.trackerMap[keyPrev]
	if ok {
		delete(t.trackerMap, keyPrev)
	}
	return tracker
}
