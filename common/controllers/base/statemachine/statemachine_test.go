// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package statemachine

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/verrazzano/verrazzano-modules/common/actionspi"
	"github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano/pkg/log/vzlog"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	vzspi "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"testing"
)

type behavior struct {
	methodName string
	visited    bool
	boolResult bool
	ctrl.Result
}

type behaviorMap map[string]*behavior

type handler struct {
	actionNeeded bool
	behaviorMap
}

const (
	getActionName               = "GetActionName"
	initFunc                    = "init"
	isActionNeeded              = "isActionNeeded"
	preActionUpdateStatus       = "preActionUpdateStatus"
	preAction                   = "preAction"
	isPreActionDone             = "isPreActionDone"
	actionUpdateStatus          = "ActionUpdateStatus"
	doAction                    = "doAction"
	isActionDone                = "isActionDone"
	postAction                  = "postAction"
	postActionUpdateStatus      = "postActionUpdateStatus"
	isPostActionDone            = "isPostActionDone"
	completedActionUpdateStatus = "CompletedActionUpdateStatus"
)

// getStatesInOrder returns the states in order of execution
func getStatesInOrder() []string {
	return []string{
		getActionName,
		initFunc,
		isActionNeeded,
		preActionUpdateStatus,
		preAction,
		isPreActionDone,
		actionUpdateStatus,
		doAction,
		isActionDone,
		postActionUpdateStatus,
		postAction,
		isPostActionDone,
		completedActionUpdateStatus,
	}
}

func getBehaviorMap() behaviorMap {
	m := make(map[string]*behavior)
	for _, s := range getStatesInOrder() {
		m[s] = &behavior{boolResult: true}
	}
	return m
}

func TestVisited(t *testing.T) {
	asserts := assert.New(t)

	h := handler{
		actionNeeded: true,
		behaviorMap:  getBehaviorMap(),
	}
	cr := &v1alpha1.ModuleLifecycle{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test",
			Namespace:  "testns",
			UID:        "uid-123",
			Generation: 1,
		},
	}
	sm := StateMachine{
		Scheme:   nil,
		CR:       cr,
		HelmInfo: &actionspi.HelmInfo{},
		Handler:  h,
	}
	ctx, err := vzspi.NewMinimalContext(nil, vzlog.DefaultLogger())
	asserts.NoError(err)

	res := sm.Execute(ctx)
	asserts.False(res.Requeue)

	// Make sure all the states were visited, check in order
	for _, s := range getStatesInOrder() {
		b := h.behaviorMap[s]
		asserts.True(b.visited, fmt.Sprintf("State %s not visited", s))
	}
}

// Implement the handler SPI
func (h handler) GetActionName() string {
	b := h.behaviorMap[getActionName]
	b.visited = true
	return "install"
}

func (h handler) Init(context spi.ComponentContext, config actionspi.HandlerConfig) (ctrl.Result, error) {
	return h.procHandlerCall(initFunc)
}

func (h handler) IsActionNeeded(context spi.ComponentContext) (bool, ctrl.Result, error) {
	return h.procHandlerBool(isActionNeeded)
}

func (h handler) PreAction(context spi.ComponentContext) (ctrl.Result, error) {
	return h.procHandlerCall(preAction)
}

func (h handler) PreActionUpdateStatus(context spi.ComponentContext) (ctrl.Result, error) {
	return h.procHandlerCall(preActionUpdateStatus)
}

func (h handler) IsPreActionDone(context spi.ComponentContext) (bool, ctrl.Result, error) {
	return h.procHandlerBool(isPreActionDone)
}

func (h handler) ActionUpdateStatus(context spi.ComponentContext) (ctrl.Result, error) {
	return h.procHandlerCall(actionUpdateStatus)
}

func (h handler) DoAction(context spi.ComponentContext) (ctrl.Result, error) {
	return h.procHandlerCall(doAction)
}

func (h handler) IsActionDone(context spi.ComponentContext) (bool, ctrl.Result, error) {
	return h.procHandlerBool(isActionDone)
}

func (h handler) PostActionUpdateStatus(context spi.ComponentContext) (ctrl.Result, error) {
	return h.procHandlerCall(postActionUpdateStatus)
}

func (h handler) PostAction(context spi.ComponentContext) (ctrl.Result, error) {
	return h.procHandlerCall(postAction)
}

func (h handler) IsPostActionDone(context spi.ComponentContext) (bool, ctrl.Result, error) {
	return h.procHandlerBool(isPostActionDone)
}

func (h handler) CompletedActionUpdateStatus(context spi.ComponentContext) (ctrl.Result, error) {
	return h.procHandlerCall(completedActionUpdateStatus)
}

func (h handler) procHandlerCall(name string) (ctrl.Result, error) {
	b := h.behaviorMap[name]
	b.visited = true
	return b.Result, nil
}

func (h handler) procHandlerBool(name string) (bool, ctrl.Result, error) {
	b := h.behaviorMap[name]
	b.visited = true
	return b.boolResult, b.Result, nil
}
