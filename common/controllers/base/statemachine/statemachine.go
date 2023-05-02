// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package statemachine

import (
	"fmt"
	compspi "github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/action_spi"
	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	vzspi "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// state identifies the state of a component during action
type state string

const (
	// stateInit is the state when a component is initialized
	stateInit state = "stateInit"

	// stateCheckActionNeeded is the state to check if action is needed
	stateCheckActionNeeded state = "stateCheckActionNeeded"

	// statePreActionUpdateStatus is the state when the status is updated to start pre action
	statePreActionUpdateStatus state = "statePreActionUpdateStatus"

	// statePreAction is the state when a component does a pre-action
	statePreAction state = "statePreAction"

	// statePreActionWaitDone is the state when a component is waiting for pre-action to be done
	statePreActionWaitDone state = "statePreActionWaitDone"

	// stateActionUpdateStatus is the state when the status is updated to start action
	stateActionUpdateStatus state = "stateActionUpdateStatus"

	// stateAction is the state where a component does an action
	stateAction state = "stateAction"

	// stateActionWaitDone is the state when a component is waiting for action to be done
	stateActionWaitDone state = "stateActionWaitDone"

	// statePostActionUpdateStatus is the state when the status is updated to start post-action
	statePostActionUpdateStatus state = "statePostActionUpdateStatus"

	// statePostAction is the state when a component does a post-action
	statePostAction state = "statePostAction"

	// statePostActionWaitDone is the state when a component is waiting for post-action to be done
	statePostActionWaitDone state = "statePostActionWaitDone"

	// stateCompleteUpdateStatus is the state when the status is updated to completed
	stateCompleteUpdateStatus state = "stateCompleteUpdateStatus"

	// stateEnd is the terminal state
	stateEnd state = "stateEnd"
)

type StateMachine struct {
	*runtime.Scheme
	CR       client.Object
	HelmInfo *compspi.HelmInfo
	Handler  compspi.LifecycleActionHandler
}

func (s StateMachine) doStateMachine(compCtx vzspi.ComponentContext) ctrl.Result {
	tracker := getTracker(s.CR.GetGeneration(), s.CR.GetUID())

	actionName := s.Handler.GetActionName()
	compContext := compCtx.Init("component").Operation(actionName)
	nsn := fmt.Sprintf("%s/%s", s.CR.GetNamespace(), s.CR.GetName())

	for s.Tracker.state != stateEnd {
		switch s.Tracker.state {
		case stateInit:
			// Init the Handler
			config := compspi.HandlerConfig{
				HelmInfo: *s.HelmInfo,
				CR:       s.CR,
				Scheme:   s.Scheme,
			}
			res, err := s.Handler.Init(compContext, config)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			s.Tracker.state = stateCheckActionNeeded

		case stateCheckActionNeeded:
			needed, res, err := s.Handler.IsActionNeeded(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			if !needed {
				s.Tracker.state = stateEnd
			} else {
				s.Tracker.state = statePreActionUpdateStatus
			}

		case statePreActionUpdateStatus:
			res, err := s.Handler.PreActionUpdateStatus(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			s.Tracker.state = statePreAction

		case statePreAction:
			compCtx.Log().Progressf("Doing pre-%s for %s", actionName, nsn)
			res, err := s.Handler.PreAction(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			s.Tracker.state = statePreActionWaitDone

		case statePreActionWaitDone:
			done, res, err := s.Handler.IsPreActionDone(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			if !done {
				return util.NewRequeueWithShortDelay()
			}
			s.Tracker.state = stateActionUpdateStatus

		case stateActionUpdateStatus:
			res, err := s.Handler.PreActionUpdateStatus(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			s.Tracker.state = stateAction

		case stateAction:
			compCtx.Log().Progressf("Doing %s for %s", actionName, nsn)
			res, err := s.Handler.DoAction(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			s.Tracker.state = stateActionWaitDone

		case stateActionWaitDone:
			done, res, err := s.Handler.IsActionDone(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			if !done {
				return util.NewRequeueWithShortDelay()
			}
			s.Tracker.state = statePostAction

		case statePostActionUpdateStatus:
			res, err := s.Handler.PostActionUpdateStatus(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			s.Tracker.state = statePostAction

		case statePostAction:
			compCtx.Log().Progressf("Doing post-%s for %s", actionName, nsn)
			res, err := s.Handler.PostAction(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			s.Tracker.state = statePostActionWaitDone

		case statePostActionWaitDone:
			done, res, err := s.Handler.IsPostActionDone(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			if !done {
				return util.NewRequeueWithShortDelay()
			}
			s.Tracker.state = stateCompleteUpdateStatus

		case stateCompleteUpdateStatus:
			res, err := s.Handler.PostActionUpdateStatus(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			s.Tracker.state = stateEnd
		}
	}
	return ctrl.Result{}
}
