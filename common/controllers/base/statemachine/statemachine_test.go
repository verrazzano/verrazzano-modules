// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package statemachine

import (
	"github.com/verrazzano/verrazzano-modules/common/actionspi"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	ctrl "sigs.k8s.io/controller-runtime"
)

type behavior struct {
	returnError bool
	visited     bool
}

type handler struct{}

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

func getBehaviors() map[string]behavior {

	return map[string]behavior{
		getActionName:          {},
		initFunc:               {},
		isActionNeeded:         {},
		preActionUpdateStatus:  {},
		preAction:              {},
		isPreActionDone:        {},
		actionUpdateStatus:     {},
		actionfunc:             {},
		isActionDone:           {},
		postActionUpdateStatus: {},
		postAction:             {},
		isPostActionDone:       {},
	}
}

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
	//TODO implement me
	panic("implement me")
}

func (h handler) CompletedActionUpdateStatus(context spi.ComponentContext) (ctrl.Result, error) {
	//TODO implement me
	panic("implement me")
}
