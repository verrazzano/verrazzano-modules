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
	done       bool
	ctrl.Result
}

type behaviorMap map[string]behavior

type handler struct {
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
	actionfunc                  = "Action"
	isActionDone                = "isActionDone"
	postAction                  = "postAction"
	postActionUpdateStatus      = "postActionUpdateStatus"
	isPostActionDone            = "isPostActionDone"
	completedActionUpdateStatus = "CompletedActionUpdateStatus"
)

func getStates() []string {
	return []string{
		getActionName,
		initFunc,
		isActionNeeded,
		preActionUpdateStatus,
		preAction,
		isPreActionDone,
		actionUpdateStatus,
		actionfunc,
		isActionDone,
		postActionUpdateStatus,
		postAction,
		isPostActionDone,
	}
}

func getBehaviorMap() behaviorMap {
	m := make(map[string]behavior)
	for _, s := range getStates() {
		m[s] = behavior{}
	}
	return m
}

func Test(t *testing.T) {
	asserts := assert.New(t)

	h := handler{
		behaviorMap: getBehaviorMap(),
	}
	cr := &v1alpha1.ModuleLifecycle{
		ObjectMeta: metav1.ObjectMeta{
			Name:       fmt.Sprintf("%s-%d", "fakeName", y),
			Namespace:  "mynamespace",
			UID:        "uid-123",
			Generation: 1,
		},
	}
	sm := StateMachine{
		Scheme:   nil,
		CR:       cr,
		HelmInfo: nil,
		Handler:  h,
	}
	ctx, err := vzspi.NewMinimalContext(nil, vzlog.DefaultLogger())
	asserts.Fail(fmt.Sprintf("Failed to get context %s", err))

	res := sm.Execute(ctx)
	asserts.False(res.Requeue)
}

// Implement the handler SPI
func (h handler) GetActionName() string {
	//TODO implement me
	panic("implement me")
}

func (h handler) Init(context spi.ComponentContext, config actionspi.HandlerConfig) (ctrl.Result, error) {
	//TODO implement me
	panic("implement me")
}

func (h handler) IsActionNeeded(context spi.ComponentContext) (bool, ctrl.Result, error) {
	//TODO implement me
	panic("implement me")
}

func (h handler) PreAction(context spi.ComponentContext) (ctrl.Result, error) {
	//TODO implement me
	panic("implement me")
}

func (h handler) PreActionUpdateStatus(context spi.ComponentContext) (ctrl.Result, error) {
	//TODO implement me
	panic("implement me")
}

func (h handler) IsPreActionDone(context spi.ComponentContext) (bool, ctrl.Result, error) {
	//TODO implement me
	panic("implement me")
}

func (h handler) ActionUpdateStatus(context spi.ComponentContext) (ctrl.Result, error) {
	//TODO implement me
	panic("implement me")
}

func (h handler) DoAction(context spi.ComponentContext) (ctrl.Result, error) {
	//TODO implement me
	panic("implement me")
}

func (h handler) IsActionDone(context spi.ComponentContext) (bool, ctrl.Result, error) {
	//TODO implement me
	panic("implement me")
}

func (h handler) PostActionUpdateStatus(context spi.ComponentContext) (ctrl.Result, error) {
	//TODO implement me
	panic("implement me")
}

func (h handler) PostAction(context spi.ComponentContext) (ctrl.Result, error) {
	//TODO implement me
	panic("implement me")
}

func (h handler) IsPostActionDone(context spi.ComponentContext) (bool, ctrl.Result, error) {
	return h.procHandlerCall(isPostActionDone)
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
	return b.done, b.Result, nil
}
