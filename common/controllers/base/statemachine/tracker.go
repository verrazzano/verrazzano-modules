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

// getTrackerKey gets the stateTracker key for the Verrazzano resource
func getTrackerKey(CR client.Object, gen int64) string {
	return fmt.Sprintf("%s-%s-%v-%s", CR.GetNamespace(), CR.GetName(), gen, string(CR.GetUID()))
}

// getTracker gets the install stateTracker for Verrazzano
func getTracker(CR client.Object, initialState state) *stateTracker {
	gen := CR.GetGeneration()
	mutex := sync.RWMutex{}
	mutex.Lock()
	defer mutex.Unlock()
	key := getTrackerKey(CR, gen)
	tracker, ok := trackerMap[key]
	// If the entry is missing then create a new entry
	if !ok {
		tracker = &stateTracker{
			state: initialState,
			gen:   CR.GetGeneration(),
		}
		trackerMap[key] = tracker

		// Delete the previous entry if it exists
		key := getTrackerKey(CR, gen-1)
		_, ok := trackerMap[key]
		if ok {
			delete(trackerMap, key)
		}
	}
	return tracker
}
