// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package lifecycle

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// stateTracker keeps an in-memory state for a component doing actions
type stateTracker struct {
	state   componentState
	gen     int64
	compMap map[string]*componentTrackerContext
}

// componentTrackerContext has the component context stateTracker
type componentTrackerContext struct {
	actionState componentState
}

// trackerMap has a map of trackers with key from VZ name, namespace, and UID
var trackerMap = make(map[string]*stateTracker)

// getTrackerKey gets the stateTracker key for the Verrazzano resource
func getTrackerKey(cr metav1.ObjectMeta) string {
	return fmt.Sprintf("%s-%s-%s", cr.Namespace, cr.Name, string(cr.UID))
}

// getTracker gets the install stateTracker for Verrazzano
func getTracker(cr metav1.ObjectMeta, initialState componentState) *stateTracker {
	key := getTrackerKey(cr)
	vuc, ok := trackerMap[key]
	// If the entry is missing or the generation is different create a new entry
	if !ok || vuc.gen != cr.Generation {
		vuc = &stateTracker{
			state:   initialState,
			gen:     cr.Generation,
			compMap: make(map[string]*componentTrackerContext),
		}
		trackerMap[key] = vuc
	}
	return vuc
}

// deleteTracker deletes the stateTracker for the Verrazzano resource
func deleteTracker(cr metav1.ObjectMeta) {
	key := getTrackerKey(cr)
	_, ok := trackerMap[key]
	if ok {
		delete(trackerMap, key)
	}
}
