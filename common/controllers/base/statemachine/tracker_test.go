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

func Test(t *testing.T) {
	tests := []struct {
		name        string
		threadCount int
		crName      string
		crCount     int
		trackerMap  map[string]*stateTracker
	}{
		{
			name:        "test-init",
			threadCount: 100,
			crName:      "test1",
			trackerMap:  nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			asserts := assert.New(t)
			if test.trackerMap != nil {
				trackerMap = test.trackerMap
			}
			asserts.NotNil(trackerMap)
			var wg sync.WaitGroup
			for i := 1; i <= test.threadCount; i++ {
				wg.Add(1)
				go func(y int) {
					defer wg.Done()
					cr := &v1alpha1.ModuleLifecycle{
						ObjectMeta: metav1.ObjectMeta{
							Name:       fmt.Sprintf("%s-%d", "fakeName", y),
							Namespace:  "mynamespace",
							UID:        "uid-123",
							Generation: 1,
						},
					}
					state := getRandomState()
					tracker := ensureTracker(cr, state)
					asserts.NotNil(tracker)
					asserts.Equal(state, tracker.state)
					key := getTrackerKey(cr, cr.Generation)
					tracker2 := getTracker(key)
					if tracker2 == nil {
						asserts.NotNil(tracker2)
					}
					asserts.NotNil(tracker2)
					asserts.Equal(tracker.state, tracker2.state)
				}(i)
			}
			wg.Wait()
		})
	}
}

// get a random state
func getRandomState() state {
	states := []state{stateInit, stateAction, statePostAction, statePreAction, stateActionUpdateStatus, stateEnd}
	return states[rand.IntnRange(0, len(states)-1)]
}
