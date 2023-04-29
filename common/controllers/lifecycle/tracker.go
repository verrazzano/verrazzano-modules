// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package lifecycle

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
func getTrackerKey(meta metav1.ObjectMeta, gen int64) string {
	return fmt.Sprintf("%s-%s-%v-%s", meta.Namespace, meta.Name, gen, string(meta.UID))
}

// getTracker gets the install stateTracker for Verrazzano
func getTracker(meta metav1.ObjectMeta, initialState state) *stateTracker {
	mutex := sync.RWMutex{}
	mutex.Lock()
	defer mutex.Unlock()
	key := getTrackerKey(meta, meta.Generation)
	tracker, ok := trackerMap[key]
	// If the entry is missing then create a new entry
	if !ok {
		tracker = &stateTracker{
			state: initialState,
			gen:   meta.Generation,
		}
		trackerMap[key] = tracker

		// Delete the previous entry if it exists
		key := getTrackerKey(meta, meta.Generation-1)
		_, ok := trackerMap[key]
		if ok {
			delete(trackerMap, key)
		}
	}
	return tracker
}
