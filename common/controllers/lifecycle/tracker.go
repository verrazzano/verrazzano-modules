// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package lifecycle

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// installTracker has the Install context for the Verrazzano Install
// This tracker keeps an in-memory Install state for Verrazzano and the components that
// are being Install.
type installTracker struct {
	state   string
	gen     int64
	compMap map[string]*componentTrackerContext
}

// installTrackerMap has a map of InstallTrackers with key from VZ name, namespace, and UID
var installTrackerMap = make(map[string]*installTracker)

// getTrackerKey gets the tracker key for the Verrazzano resource
func getTrackerKey(cr metav1.ObjectMeta) string {
	return fmt.Sprintf("%s-%s-%s", cr.Namespace, cr.Name, string(cr.UID))
}

// getInstallTracker gets the install tracker for Verrazzano
func getInstallTracker(cr metav1.ObjectMeta, initialState string) *installTracker {
	key := getTrackerKey(cr)
	vuc, ok := installTrackerMap[key]
	// If the entry is missing or the generation is different create a new entry
	if !ok || vuc.gen != cr.Generation {
		vuc = &installTracker{
			state:   initialState,
			gen:     cr.Generation,
			compMap: make(map[string]*componentTrackerContext),
		}
		installTrackerMap[key] = vuc
	}
	return vuc
}

// deleteInstallTracker deletes the install tracker for the Verrazzano resource
func deleteInstallTracker(cr metav1.ObjectMeta) {
	key := getTrackerKey(cr)
	_, ok := installTrackerMap[key]
	if ok {
		delete(installTrackerMap, key)
	}
}
