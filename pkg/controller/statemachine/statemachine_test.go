// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package statemachine

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/result"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/spi/handlerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

type behavior struct {
	methodName string
	visited    bool
	boolResult bool
	result.Result
}

type behaviorMap map[string]*behavior

type handler struct {
	workNeeded bool
	behaviorMap
}

const (
	getWorkName               = "GetWorkName"
	isWorkNeeded              = "isWorkNeeded"
	checkDependencies         = "checkDependencies"
	preWorkUpdateStatus       = "preWorkUpdateStatus"
	preWork                   = "preWork"
	workUpdateStatus          = "DoWorkUpdateStatus"
	doWork                    = "doWork"
	isWorkDone                = "isWorkDone"
	postWork                  = "postWork"
	postWorkUpdateStatus      = "postWorkUpdateStatus"
	completedWorkUpdateStatus = "WorkCompletedUpdateStatus"
)

// getStatesInOrder returns the states in order of execution
func getStatesInOrder() []string {
	return []string{
		getWorkName,
		isWorkNeeded,
		checkDependencies,
		preWorkUpdateStatus,
		preWork,
		workUpdateStatus,
		doWork,
		isWorkDone,
		postWorkUpdateStatus,
		postWork,
		completedWorkUpdateStatus,
	}
}

// TestAllStatesSucceed tests that all the states are visited
// GIVEN a state machine
// WHEN the state machine is executed and all handler methods return success
// THEN ensure that every state in the state machine is executed
func TestAllStatesSucceed(t *testing.T) {
	asserts := assert.New(t)

	h := handler{
		workNeeded:  true,
		behaviorMap: getBehaviorMap(),
	}
	cr := &v1alpha1.Module{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "TestAllStatesSucceed",
			Namespace:  "testns",
			UID:        "uid-123",
			Generation: 1,
		},
	}
	sm := StateMachine{
		CR:      cr,
		Handler: h,
	}
	ctx := handlerspi.HandlerContext{Client: nil, Log: vzlog.DefaultLogger(), HelmInfo: handlerspi.HelmInfo{}}

	res := sm.Execute(ctx)
	asserts.False(res.ShouldRequeue())

	// Make sure all the states were visited, check in order
	for _, s := range getStatesInOrder() {
		b := h.behaviorMap[s]
		asserts.True(b.visited, fmt.Sprintf("State %s not visited", s))
	}
}

// TestEachStateRequeue tests that all the states handle requeue
// GIVEN multiple state machines
// WHEN each state machine is executed from the beginning, and every state is tested to return requeue
// THEN ensure that each state before the state that returns requeue result is executed,
// AND each subsequent state is not executed
func TestEachStateRequeue(t *testing.T) {
	asserts := assert.New(t)
	ctx := handlerspi.HandlerContext{Client: nil, Log: vzlog.DefaultLogger(), HelmInfo: handlerspi.HelmInfo{}}

	// Check for a Requeue result for every state
	for i, s := range getStatesInOrder() {
		if s == getWorkName {
			continue
		}

		h := handler{
			workNeeded:  true,
			behaviorMap: getBehaviorMap(),
		}
		cr := &v1alpha1.Module{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "TestEachStateRequeue",
				Namespace:  "testns",
				UID:        "uid-123",
				Generation: 1 + int64(i),
			},
		}
		sm := StateMachine{
			CR:      cr,
			Handler: h,
		}
		b := h.behaviorMap[s]
		b.Result = result.NewResultShortRequeueDelay()

		res := sm.Execute(ctx)
		asserts.True(res.ShouldRequeue())

		// Make sure all the states were visited only up to the one that requeued, check in order
		expectedVisited := true
		for _, s := range getStatesInOrder() {
			if s == getWorkName {
				continue
			}
			b := h.behaviorMap[s]
			asserts.Equal(expectedVisited, b.visited, fmt.Sprintf("State %s visited wrong, should be %v", s, b.visited))
			if b.ShouldRequeue() {
				expectedVisited = false
			}
		}
	}
}

// TestNotDone tests that all the states that check for done are working correctly
// GIVEN multiple state machines
// WHEN each state machine is executed from the beginning
// THEN ensure that each state that checks for done returns a requeue when the handler returns `not done`
// AND ensure that each state before the requeue state is executed
// AND each subsequent state is not executed
func TestNotDone(t *testing.T) {
	asserts := assert.New(t)
	ctx := handlerspi.HandlerContext{Client: nil, Log: vzlog.DefaultLogger(), HelmInfo: handlerspi.HelmInfo{}}

	tests := []struct {
		name          string
		stateNotDone  string
		expectRequeue bool
	}{
		{
			name:          "isWorkNeeded",
			stateNotDone:  isWorkNeeded,
			expectRequeue: false,
		},
		{
			name:          "isWorkDone",
			stateNotDone:  isWorkDone,
			expectRequeue: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := handler{
				workNeeded:  true,
				behaviorMap: getBehaviorMap(),
			}
			cr := &v1alpha1.Module{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "TestEachStateRequeue-" + test.stateNotDone,
					Namespace:  "testns",
					UID:        "uid-123",
					Generation: 1,
				},
			}
			sm := StateMachine{
				CR:      cr,
				Handler: h,
			}
			b := h.behaviorMap[test.stateNotDone]
			b.boolResult = false

			res := sm.Execute(ctx)
			asserts.Equal(test.expectRequeue, res.ShouldRequeue(), fmt.Sprintf("State %s should cause requeue", test.stateNotDone))

			// Make sure all the states were visited only up to the one that requeued, check in order
			expectedVisited := true
			for _, s := range getStatesInOrder() {
				if s == getWorkName {
					continue
				}
				b := h.behaviorMap[s]
				asserts.Equal(expectedVisited, b.visited, fmt.Sprintf("State %s visited wrong, should be %v", s, b.visited))
				if !b.boolResult {
					expectedVisited = false
				}
			}
		})
	}
}

func getBehaviorMap() behaviorMap {
	m := make(map[string]*behavior)
	for _, s := range getStatesInOrder() {
		m[s] = &behavior{boolResult: true, methodName: s, Result: result.NewResult()}
	}
	return m
}

// Implement the handler SPI
func (h handler) GetWorkName() string {
	b := h.behaviorMap[getWorkName]
	b.visited = true
	return "install"
}

func (h handler) IsWorkNeeded(context handlerspi.HandlerContext) (bool, result.Result) {
	return h.procHandlerBool(isWorkNeeded)
}

func (h handler) CheckDependencies(context handlerspi.HandlerContext) result.Result {
	return h.procHandlerCall(checkDependencies)
}

func (h handler) PreWork(context handlerspi.HandlerContext) result.Result {
	return h.procHandlerCall(preWork)
}

func (h handler) PreWorkUpdateStatus(context handlerspi.HandlerContext) result.Result {
	return h.procHandlerCall(preWorkUpdateStatus)
}

func (h handler) DoWorkUpdateStatus(context handlerspi.HandlerContext) result.Result {
	return h.procHandlerCall(workUpdateStatus)
}

func (h handler) DoWork(context handlerspi.HandlerContext) result.Result {
	return h.procHandlerCall(doWork)
}

func (h handler) IsWorkDone(context handlerspi.HandlerContext) (bool, result.Result) {
	return h.procHandlerBool(isWorkDone)
}

func (h handler) PostWorkUpdateStatus(context handlerspi.HandlerContext) result.Result {
	return h.procHandlerCall(postWorkUpdateStatus)
}

func (h handler) PostWork(context handlerspi.HandlerContext) result.Result {
	return h.procHandlerCall(postWork)
}

func (h handler) WorkCompletedUpdateStatus(context handlerspi.HandlerContext) result.Result {
	return h.procHandlerCall(completedWorkUpdateStatus)
}

func (h handler) procHandlerCall(name string) result.Result {
	b := h.behaviorMap[name]
	b.visited = true
	return b.Result
}

func (h handler) procHandlerBool(name string) (bool, result.Result) {
	b := h.behaviorMap[name]
	b.visited = true
	return b.boolResult, b.Result
}
