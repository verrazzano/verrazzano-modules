// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package statemachine

import (
	"fmt"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

// stateTracker keeps an in-memory state for a ModuleCR execute the state machine.
type stateTracker struct {
	state state
	gen   int64
	key   string
}

// trackerMap has a map of trackers with key from VZ name, namespace, and UID.
var trackerMap = make(map[string]*stateTracker)

// trackerMutex is used to access the map concurrently.
var trackerMutex sync.RWMutex

// getTrackerKey gets the stateTracker key for a ModuleCR.
func getTrackerKey(CR client.Object) string {
	return fmt.Sprintf("%s-%s-%s", CR.GetNamespace(), CR.GetName(), string(CR.GetUID()))
}

// ensureTracker gets the stateTracker, creating a new one if needed.
func ensureTracker(CR client.Object, initialState state) *stateTracker {
	trackerMutex.Lock()
	defer trackerMutex.Unlock()
	key := getTrackerKey(CR)
	tracker, ok := trackerMap[key]
	if ok {
		// The CR generation must match the tracker
		if CR.GetGeneration() == tracker.gen {
			return tracker
		}
		delete(trackerMap, key)
		vzlog.DefaultLogger().Debugf("Deleting tracker for %s, generation %s", key, CR.GetGeneration())
	}

	// create a new tracker and save it in the map
	vzlog.DefaultLogger().Debugf("Creating a tracker for key: %s, generation %s", key, CR.GetGeneration())
	tracker = &stateTracker{
		state: initialState,
		gen:   CR.GetGeneration(),
		key:   key,
	}
	trackerMap[key] = tracker
	return tracker
}

// DeleteTracker deletes the tracker for the given CR
func DeleteTracker(CR client.Object) {
	trackerMutex.Lock()
	defer trackerMutex.Unlock()
	key := getTrackerKey(CR)
	_, ok := trackerMap[key]
	if ok {
		delete(trackerMap, key)
		vzlog.DefaultLogger().Debugf("Deleted tracker for %s", key, CR.GetGeneration())
	}
}
