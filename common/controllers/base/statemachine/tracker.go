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

// trackerMap has a map of trackers with key from VZ name, namespace, and UID
var trackerMap = make(map[string]*stateTracker)

// trackerMutex is used to access the map concurrently
var trackerMutex = sync.RWMutex{}

// getTrackerKey gets the stateTracker key for the Verrazzano resource
func getTrackerKey(CR client.Object, gen int64) string {
	return fmt.Sprintf("%s-%s-%v-%s", CR.GetNamespace(), CR.GetName(), gen, string(CR.GetUID()))
}

// ensureTracker gets the stateTracker, creating a new one if needed
func ensureTracker(CR client.Object, initialState state) *stateTracker {
	key := getTrackerKey(CR, CR.GetGeneration())
	tracker := getTracker(key)
	if tracker == nil {
		tracker = createTracker(CR, initialState, key)
	}
	return tracker
}

// getTracker gets the stateTracker
func getTracker(key string) *stateTracker {
	trackerMutex.RLock()
	defer trackerMutex.RUnlock()
	tracker, _ := trackerMap[key]
	return tracker
}

// createTracker creates a stateTracker
func createTracker(CR client.Object, initialState state, key string) *stateTracker {
	trackerMutex.Lock()
	defer trackerMutex.Unlock()
	tracker := &stateTracker{
		state: initialState,
		gen:   CR.GetGeneration(),
	}
	trackerMap[key] = tracker

	// Delete the previous entry if it exists
	if CR.GetGeneration() == 1 {
		return tracker
	}
	keyPrev := getTrackerKey(CR, CR.GetGeneration()-1)
	_, ok := trackerMap[keyPrev]
	if ok {
		delete(trackerMap, keyPrev)
	}
	return tracker
}
