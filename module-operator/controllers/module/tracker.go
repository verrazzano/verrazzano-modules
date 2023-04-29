// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package module

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
func getTrackerKey(cr metav1.ObjectMeta) string {
	return fmt.Sprintf("%s-%s-%v-%s", cr.Namespace, cr.Name, cr.Generation, string(cr.UID))
}

// getTracker gets the install stateTracker for Verrazzano
func getTracker(cr metav1.ObjectMeta, initialState state) *stateTracker {
	mutex := sync.RWMutex{}
	mutex.Lock()
	defer mutex.Unlock()
	key := getTrackerKey(cr)
	vuc, ok := trackerMap[key]
	// If the entry is missing then create a new entry
	if !ok {
		vuc = &stateTracker{
			state: initialState,
			gen:   cr.Generation,
		}
		trackerMap[key] = vuc
	}
	return vuc
}
