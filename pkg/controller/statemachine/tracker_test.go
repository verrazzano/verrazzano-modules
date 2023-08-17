// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package statemachine

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"sync"
	"testing"
)

// TestEnsureTracker tests the ensureTracker method
// GIVEN a trackerContext
// WHEN multiple CRs are tracked concurrently
// THEN ensure that the correct tracker is returned
func TestEnsureTracker(t *testing.T) {
	asserts := assert.New(t)
	startingMapSize := len(trackerMap)

	const threadCount = 1000
	var wg sync.WaitGroup
	for i := 1; i <= threadCount; i++ {
		wg.Add(1)
		go func(y int) {
			defer wg.Done()
			cr := &v1alpha1.Module{
				ObjectMeta: metav1.ObjectMeta{
					Name:       fmt.Sprintf("%s-%d", "fakeName", y),
					Namespace:  "mynamespace",
					UID:        "uid-123",
					Generation: 1,
				},
			}
			// Ensure tracker starts at random state
			state := getRandomState()

			// Simulate reconcile loop staying at state
			loop := 1
			for loop < 5 {
				loop++
				tracker := ensureTracker(cr, state)
				asserts.NotNil(tracker)
				asserts.Equal(state, tracker.state)

				// Get another tracker, should match the first
				tracker2 := ensureTracker(cr, state)
				if tracker2 == nil {
					asserts.NotNil(tracker2)
				}
				asserts.NotNil(tracker2)
				asserts.Equal(tracker.state, tracker2.state)
			}

			// Simulate reconcile loop changing state
			loop = 1
			for loop < 5 {
				loop++
				tracker := ensureTracker(cr, state)
				asserts.NotNil(tracker)
				asserts.Equal(state, tracker.state)

				// update the state and make sure the new call to ensure tracker returns a tracker with correct state
				state = getRandomState()
				tracker.state = state
				tracker3 := ensureTracker(cr, state)
				asserts.Equal(tracker.state, tracker3.state)
			}
		}(i)
	}
	wg.Wait()
	asserts.Equal(threadCount+startingMapSize, len(trackerMap))
}

// TestNewGeneration tests that old trackers are deleted when the generation changes
// GIVEN a trackerContext
// WHEN multiple CRs are tracked concurrently and the generation changes
// THEN ensure that trackers tracking old generations of CRs are deleted
func TestNewGeneration(t *testing.T) {
	asserts := assert.New(t)
	startingMapSize := len(trackerMap)

	const threadCount = 1000
	var wg sync.WaitGroup
	for i := 1; i <= threadCount; i++ {
		wg.Add(1)
		go func(y int) {
			defer wg.Done()
			cr := &v1alpha1.Module{
				ObjectMeta: metav1.ObjectMeta{
					Name:       fmt.Sprintf("%s-%d", "fakeName2", y),
					Namespace:  "mynamespace",
					UID:        "uid-12345",
					Generation: 1,
				},
			}
			// Ensure tracker starts at state
			startstate := getRandomState()
			tracker := ensureTracker(cr, startstate)
			asserts.NotNil(tracker)
			asserts.Equal(startstate, tracker.state)

			// Increment generation and get tracker again
			cr.Generation++
			tracker2 := ensureTracker(cr, state("fakestate"))

			// state and gen should never match
			asserts.NotEqual(tracker.state, tracker2.state)
			asserts.NotEqual(tracker.gen, tracker2.gen)
		}(i)
	}
	wg.Wait()

	// Assert the old trackers were removed and map size stayed at the thread count
	asserts.Equal(threadCount+startingMapSize, len(trackerMap))
}

// TestDeleteTracker tests that old trackers are deleted when DeleteTracker is called
// GIVEN a trackerContext
// WHEN multiple CRs are tracked concurrently and a tracker is deleted
// THEN ensure that only the tracker for the CR is deleted.
func TestDeleteTracker(t *testing.T) {
	asserts := assert.New(t)

	const threadCount = 1000
	var wg sync.WaitGroup
	for i := 1; i <= threadCount; i++ {
		wg.Add(1)
		go func(y int) {
			defer wg.Done()

			// Ensure tracker starts at state
			startstate := getRandomState()
			cr := buildCR(y)
			tracker := ensureTracker(cr, startstate)
			asserts.NotNil(tracker)
			asserts.Equal(startstate, tracker.state)

			// Delete the tracker if suffix is even
			if y%2 == 0 {
				DeleteTracker(cr)
			}
		}(i)
	}
	wg.Wait()

	// Make sure correct trackers were deleted
	trackerMutex.Lock()
	defer trackerMutex.Unlock()
	for i := 1; i <= threadCount; i++ {
		cr := buildCR(i)
		key := getTrackerKey(cr)
		_, ok := trackerMap[key]
		if i%2 == 0 {
			asserts.False(ok)
		} else {
			asserts.True(ok)
		}
	}
}

// get a random state
func getRandomState() state {
	states := []state{stateInit, stateWork, statePostWork, statePreWork, stateWorkUpdateStatus, stateEnd}
	return states[rand.IntnRange(0, len(states)-1)]
}

func buildCR(suffix int) *v1alpha1.Module {
	cr := &v1alpha1.Module{
		ObjectMeta: metav1.ObjectMeta{
			Name:       fmt.Sprintf("%s-%d", "fakeName2", suffix),
			Namespace:  "TestDeleteTracker",
			UID:        "uid-12345",
			Generation: 1,
		},
	}
	return cr
}
