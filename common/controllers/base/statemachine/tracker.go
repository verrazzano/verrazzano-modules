// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package statemachine

import (
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

// statracker keeps an in-memory state for a component doing actions
type statracker struct {
	state state
	gen   int64
	key   string
}

// trackerMap has a map of trackers with key from VZ name, namespace, and UID
var trackerMap = make(map[string]*statracker)

// trackerMutex is used to access the map concurrently
var trackerMutex sync.RWMutex

// getTrackerKey gets the statracker key for the Verrazzano resource
func getTrackerKey(CR client.Object, gen int64) string {
	return fmt.Sprintf("%s-%s-%v-%s", CR.GetNamespace(), CR.GetName(), gen, string(CR.GetUID()))
}

// ensureTracker gets the statracker, creating a new one if needed
func ensureTracker(CR client.Object, initialState state) *statracker {
	trackerMutex.Lock()
	defer trackerMutex.Unlock()
	key := getTrackerKey(CR, CR.GetGeneration())
	tracker, ok := trackerMap[key]
	if ok {
		return tracker
	}

	// create a new tracker and save it in the map
	tracker = &statracker{
		state: initialState,
		gen:   CR.GetGeneration(),
		key:   key,
	}
	trackerMap[key] = tracker

	// Delete the previous entry if it exists
	if CR.GetGeneration() == 1 {
		return tracker
	}
	keyPrev := getTrackerKey(CR, CR.GetGeneration()-1)
	_, ok = trackerMap[keyPrev]
	if ok {
		delete(trackerMap, keyPrev)
	}
	return tracker
}
