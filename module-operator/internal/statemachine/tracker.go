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
func getTrackerKey(CR client.Object, gen int64) string {
	return fmt.Sprintf("%s-%s-%s-gen-%v", CR.GetNamespace(), CR.GetName(), string(CR.GetUID()), gen)
}

// ensureTracker gets the stateTracker, creating a new one if needed.
func ensureTracker(CR client.Object, initialState state) *stateTracker {
	trackerMutex.Lock()
	defer trackerMutex.Unlock()
	key := getTrackerKey(CR, CR.GetGeneration())
	tracker, ok := trackerMap[key]
	if ok {
		return tracker
	}

	// create a new tracker and save it in the map
	tracker = &stateTracker{
		state: initialState,
		gen:   CR.GetGeneration(),
		key:   key,
	}
	trackerMap[key] = tracker

	vzlog.DefaultLogger().Debugf("Creating a tracker for key: %s ", key)
	// Delete the previous entry if it exists
	if CR.GetGeneration() == 1 {
		return tracker
	}
	keyPrev := getTrackerKey(CR, CR.GetGeneration()-1)
	_, ok = trackerMap[keyPrev]
	if ok {
		delete(trackerMap, keyPrev)
		vzlog.DefaultLogger().Debugf("Deleting a tracker for key: %s ", keyPrev)

	}
	return tracker
}