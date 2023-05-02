// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package modulelifecycle

import (
<<<<<<< HEAD
	"github.com/verrazzano/verrazzano-modules/common/actionspi"
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/spi"
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/statemachine"
=======
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/spi"
	compspi "github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/action_spi"
>>>>>>> 1144059a8a1b459499f8010e4ac6a7388d711f85
	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	"github.com/verrazzano/verrazzano-modules/common/pkg/k8s"
	moduleplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	vzspi "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

<<<<<<< HEAD
=======
// state identifies the state of a component during handler
type state string

const (
	// stateInit is the state when a component is initialized
	stateInit state = "stateInit"

	// stateCheckActionNeeded is the state to check if handler is needed
	stateCheckActionNeeded state = "stateCheckActionNeeded"

	// stateActionNotNeededUpdate is the state when the status is updated to not needed
	stateActionNotNeededUpdate state = "stateActionNotNeededUpdate"

	// stateStartPreActionUpdate is the state when the status is updated to start pre handler
	stateStartPreActionUpdate state = "stateStartPreActionUpdate"

	// stateStartActionUpdate is the state when the status is updated to start handler
	stateStartActionUpdate state = "stateStartActionUpdate"

	// statePreAction is the state when a component does a pre-handler
	statePreAction state = "statePreAction"

	// statePreActionWaitDone is the state when a component is waiting for pre-handler to be done
	statePreActionWaitDone state = "statePreActionWaitDone"

	// stateAction is the state where a component does an handler
	stateAction state = "stateAction"

	// stateActionWaitDone is the state when a component is waiting for handler to be done
	stateActionWaitDone state = "stateActionWaitDone"

	// statePostAction is the state when a component does a post-handler
	statePostAction state = "statePostAction"

	// statePostActionWaitDone is the state when a component is waiting for post-handler to be done
	statePostActionWaitDone state = "statePostActionWaitDone"

	// stateCompleteUpdate is the state when the status is updated to completed
	stateCompleteUpdate state = "stateCompleteUpdate"

	// stateEnd is the terminal state
	stateEnd state = "stateEnd"
)

type stateMachineContext struct {
	cr       *moduleplatform.ModuleLifecycle
	tracker  *stateTracker
	helmInfo *compspi.HelmInfo
	handler  compspi.LifecycleActionHandler
}

>>>>>>> 1144059a8a1b459499f8010e4ac6a7388d711f85
// Reconcile reconciles the ModuleLifecycle CR
func (r Reconciler) Reconcile(spictx spi.ReconcileContext, u *unstructured.Unstructured) (ctrl.Result, error) {
	cr := &moduleplatform.ModuleLifecycle{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, cr); err != nil {
		return ctrl.Result{}, err
	}
	nsn := k8s.GetNamespacedName(cr.ObjectMeta)

	// This is an imperative command, don't rerun it
	if cr.Status.State == moduleplatform.StateCompleted || cr.Status.State == moduleplatform.StateNotNeeded {
		spictx.Log.Oncef("Resource %v has already been processed, nothing to do", nsn)
		return ctrl.Result{}, nil
	}

	ctx, err := vzspi.NewMinimalContext(r.Client, spictx.Log)
	if err != nil {
		return util.NewRequeueWithShortDelay(), err
	}

	if cr.Generation == cr.Status.ObservedGeneration {
		spictx.Log.Debugf("Skipping reconcile for %v, observed generation has not change", nsn)
		return util.NewRequeueWithShortDelay(), err
	}

	helmInfo := loadHelmInfo(cr)
<<<<<<< HEAD
	handler := r.getActionHandler(cr.Spec.Action)
	if handler == nil {
=======
	tracker := getTracker(cr.ObjectMeta, stateInit)

	action := r.getAction(cr.Spec.Action)
	if action == nil {
>>>>>>> 1144059a8a1b459499f8010e4ac6a7388d711f85
		spictx.Log.Errorf("Invalid ModuleLifecycle CR handler %s", cr.Spec.Action)
		// Dont requeue, this is a fatal error
		return ctrl.Result{}, nil
	}

<<<<<<< HEAD
	sm := statemachine.StateMachine{
		Scheme:   r.Scheme,
		CR:       cr,
		HelmInfo: &helmInfo,
		Handler:  handler,
	}

	res := sm.Execute(ctx)
	return res, nil
}

func loadHelmInfo(cr *moduleplatform.ModuleLifecycle) actionspi.HelmInfo {
	helmInfo := actionspi.HelmInfo{
=======
	smc := stateMachineContext{
		cr:       cr,
		tracker:  tracker,
		helmInfo: &helmInfo,
		handler:  action,
	}

	res := r.doStateMachine(ctx, smc)
	return res, nil
}

func (r *Reconciler) doStateMachine(spiCtx vzspi.ComponentContext, s stateMachineContext) ctrl.Result {
	actionName := s.cr.Spec.Action
	compContext := spiCtx.Init("component").Operation(string(actionName))
	nsn := k8s.GetNamespacedNameString(s.cr.ObjectMeta)

	for s.tracker.state != stateEnd {
		switch s.tracker.state {
		case stateInit:
			config := compspi.HandlerConfig{
				HelmInfo: *s.helmInfo,
				CR:       s.cr,
				Scheme:   r.Scheme,
			}
			res, err := s.handler.Init(compContext, config)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			s.tracker.state = stateCheckActionNeeded

		case stateCheckActionNeeded:
			needed, res, err := s.handler.IsActionNeeded(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			if needed {
				s.tracker.state = stateStartPreActionUpdate
			} else {
				s.tracker.state = stateActionNotNeededUpdate
			}

		case stateActionNotNeededUpdate:
			cond := s.handler.GetStatusConditions().NotNeeded
			if err := UpdateStatus(r.Client, s.cr, string(cond), cond); err != nil {
				return util.NewRequeueWithShortDelay()
			}
			s.tracker.state = stateEnd

		case stateStartPreActionUpdate:
			cond := s.handler.GetStatusConditions().PreAction
			if err := UpdateStatus(r.Client, s.cr, string(cond), cond); err != nil {
				return util.NewRequeueWithShortDelay()
			}
			s.tracker.state = statePreAction

		case statePreAction:
			spiCtx.Log().Progressf("Doing pre-%s for %s", actionName, nsn)
			res, err := s.handler.PreAction(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			s.tracker.state = statePreActionWaitDone

		case statePreActionWaitDone:
			done, res, err := s.handler.IsPreActionDone(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			if !done {
				return util.NewRequeueWithShortDelay()
			}
			s.tracker.state = stateStartActionUpdate

		case stateStartActionUpdate:
			cond := s.handler.GetStatusConditions().DoAction
			if err := UpdateStatus(r.Client, s.cr, string(cond), cond); err != nil {
				return util.NewRequeueWithShortDelay()
			}
			s.tracker.state = stateAction

		case stateAction:
			spiCtx.Log().Progressf("Doing %s for %s", actionName, nsn)
			res, err := s.handler.DoAction(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			s.tracker.state = stateActionWaitDone

		case stateActionWaitDone:
			done, res, err := s.handler.IsActionDone(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			if !done {
				return util.NewRequeueWithShortDelay()
			}
			s.tracker.state = statePostAction

		case statePostAction:
			spiCtx.Log().Progressf("Doing post-%s for %s", actionName, nsn)
			res, err := s.handler.PostAction(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			s.tracker.state = statePostActionWaitDone

		case statePostActionWaitDone:
			done, res, err := s.handler.IsPostActionDone(compContext)
			if res2 := util.DeriveResult(res, err); res2.Requeue {
				return res2
			}
			if !done {
				return util.NewRequeueWithShortDelay()
			}
			s.tracker.state = stateCompleteUpdate

		case stateCompleteUpdate:
			cond := s.handler.GetStatusConditions().Completed
			if err := UpdateStatus(r.Client, s.cr, string(cond), cond); err != nil {
				return util.NewRequeueWithShortDelay()
			}
			spiCtx.Log().Progressf("Successfully completed %s for %s", actionName, nsn)

			s.tracker.state = stateEnd
		}
	}
	return ctrl.Result{}
}

func loadHelmInfo(cr *moduleplatform.ModuleLifecycle) compspi.HelmInfo {
	helmInfo := compspi.HelmInfo{
>>>>>>> 1144059a8a1b459499f8010e4ac6a7388d711f85
		HelmRelease: cr.Spec.Installer.HelmRelease,
	}
	return helmInfo
}

<<<<<<< HEAD
func (r *Reconciler) getActionHandler(action moduleplatform.ActionType) actionspi.LifecycleActionHandler {
=======
func (r *Reconciler) getAction(action moduleplatform.ActionType) compspi.LifecycleActionHandler {
>>>>>>> 1144059a8a1b459499f8010e4ac6a7388d711f85
	switch action {
	case moduleplatform.InstallAction:
		return r.comp.InstallAction
	case moduleplatform.UninstallAction:
		return r.comp.UninstallAction
	case moduleplatform.UpdateAction:
		return r.comp.UpdateAction
	case moduleplatform.UpgradeAction:
		return r.comp.UpgradeAction
	}
	return nil
}
