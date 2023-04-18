// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package lifecycle

import (
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/spi"
	compspi "github.com/verrazzano/verrazzano-modules/common/helm_component/action_spi"
	"github.com/verrazzano/verrazzano-modules/common/pkg/controllerutils"
	"github.com/verrazzano/verrazzano-modules/common/pkg/k8s"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	moduleplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"

	vzctrl "github.com/verrazzano/verrazzano-modules/module-operator/pkg/controller"
	vzspi "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
)

// componentState identifies the state of a component during action
type componentState string

const (
	// stateInit is the state when a component is initialized
	stateInit componentState = "componentStateInit"

	// stateCheckActionNeeded is the state to check if action is needed
	stateCheckActionNeeded componentState = "stateCheckActionNeeded"

	// stateActionNotNeededUpdate is the state when the status is updated to not needed
	stateActionNotNeededUpdate componentState = "stateActionNotNeededUpdate"

	// stateStartPreActionUpdate is the state when the status is updated to start pre action
	stateStartPreActionUpdate componentState = "stateStartPreActionUpdate"

	// stateStartActionUpdate is the state when the status is updated to start action
	stateStartActionUpdate componentState = "stateStartActionUpdate"

	// statePreAction is the state when a component does a pre-action
	statePreAction componentState = "statePreAction"

	// statePreActionWaitDone is the state when a component is waiting for pre-action to be done
	statePreActionWaitDone componentState = "statePreActionWaitDone"

	// stateAction is the state where a component does an action
	stateAction componentState = "stateAction"

	// stateActionWaitDone is the state when a component is waiting for action to be done
	stateActionWaitDone componentState = "stateActionWaitDone"

	// statePostAction is the state when a component does a post-action
	statePostAction componentState = "statePostAction"

	// statePostActionWaitDone is the state when a component is waiting for post-action to be done
	statePostActionWaitDone componentState = "statePostActionWaitDone"

	// stateCompleteUpdate is the state when the status is updated to completed
	stateCompleteUpdate componentState = "stateCompleteUpdate"

	// stateEnd is the terminal state
	stateEnd componentState = "stateEnd"
)

type stateMachineContext struct {
	cr        *moduleplatform.ModuleLifecycle
	tracker   *stateTracker
	chartInfo *compspi.HelmInfo
	action    compspi.LifecycleActionHandler
}

// Reconcile updates the Certificate
func (r Reconciler) Reconcile(spictx spi.ReconcileContext, u *unstructured.Unstructured) (ctrl.Result, error) {
	cr := &moduleplatform.ModuleLifecycle{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, cr); err != nil {
		return ctrl.Result{}, err
	}
	nsn := k8s.GetNamespacedName(cr.ObjectMeta)

	// This is an imperative command, don't rerun it
	if cr.Status.State == moduleplatform.StateReady || cr.Status.State == moduleplatform.StateIgnored {
		spictx.Log.Oncef("Resource %v has already been processed, nothing to do", nsn)
		return ctrl.Result{}, nil
	}

	ctx, err := vzspi.NewMinimalContext(r.Client, spictx.Log)
	if err != nil {
		return controllerutils.NewRequeueWithShortDelay(), err
	}

	if cr.Generation == cr.Status.ObservedGeneration {
		spictx.Log.Debugf("Skipping reconcile for %v, observed generation has not change", nsn)
		return controllerutils.NewRequeueWithShortDelay(), err
	}

	helmInfo := loadHelmInfo(cr)
	tracker := getTracker(cr.ObjectMeta, stateInit)

	action := r.getAction(cr.Spec.Action)
	if action == nil {
		spictx.Log.Errorf("Invalid ModuleLifecycle CR action %s", cr.Spec.Action)
		// Dont requeue, this is a fatal error
		return ctrl.Result{}, nil
	}

	smc := stateMachineContext{
		cr:        cr,
		tracker:   tracker,
		chartInfo: &helmInfo,
		action:    action,
	}

	res := r.doStateMachine(ctx, smc)
	return res, nil
}

func (r *Reconciler) doStateMachine(spiCtx vzspi.ComponentContext, s stateMachineContext) ctrl.Result {
	compContext := spiCtx.Init("component").Operation(string(s.cr.Spec.Action))

	for s.tracker.state != stateEnd {
		switch s.tracker.state {
		case stateInit:
			res, err := s.action.Init(compContext, s.chartInfo)
			if res2 := procResult(res, err); res2.Requeue {
				return res2
			}
			s.tracker.state = stateCheckActionNeeded

		case stateCheckActionNeeded:
			needed, res, err := s.action.IsActionNeeded(compContext)
			if res2 := procResult(res, err); res2.Requeue {
				return res2
			}
			if needed {
				s.tracker.state = stateStartPreActionUpdate
			} else {
				s.tracker.state = stateActionNotNeededUpdate
			}

		case stateActionNotNeededUpdate:
			cond := s.action.GetStatusConditions().NotNeeded
			if err := UpdateStatus(r.Client, s.cr, string(cond), cond); err != nil {
				return controllerutils.NewRequeueWithShortDelay()
			}
			s.tracker.state = stateEnd

		case stateStartPreActionUpdate:
			cond := s.action.GetStatusConditions().PreAction
			if err := UpdateStatus(r.Client, s.cr, string(cond), cond); err != nil {
				return controllerutils.NewRequeueWithShortDelay()
			}
			s.tracker.state = statePreAction

		case statePreAction:
			res, err := s.action.PreAction(compContext)
			if res2 := procResult(res, err); res2.Requeue {
				return res2
			}
			s.tracker.state = statePreActionWaitDone

		case statePreActionWaitDone:
			done, res, err := s.action.IsPreActionDone(compContext)
			if res2 := procResult(res, err); res2.Requeue {
				return res2
			}
			if !done {
				return controllerutils.NewRequeueWithShortDelay()
			}
			s.tracker.state = stateStartActionUpdate

		case stateStartActionUpdate:
			cond := s.action.GetStatusConditions().DoAction
			if err := UpdateStatus(r.Client, s.cr, string(cond), cond); err != nil {
				return controllerutils.NewRequeueWithShortDelay()
			}
			s.tracker.state = stateAction

		case stateAction:
			res, err := s.action.DoAction(compContext)
			if res2 := procResult(res, err); res2.Requeue {
				return res2
			}
			s.tracker.state = stateActionWaitDone

		case stateActionWaitDone:
			done, res, err := s.action.IsActionDone(compContext)
			if res2 := procResult(res, err); res2.Requeue {
				return res2
			}
			if !done {
				return controllerutils.NewRequeueWithShortDelay()
			}
			s.tracker.state = statePostAction

		case statePostAction:
			res, err := s.action.PostAction(compContext)
			if res2 := procResult(res, err); res2.Requeue {
				return res2
			}
			s.tracker.state = statePostActionWaitDone

		case statePostActionWaitDone:
			done, res, err := s.action.IsPostActionDone(compContext)
			if res2 := procResult(res, err); res2.Requeue {
				return res2
			}
			if !done {
				return controllerutils.NewRequeueWithShortDelay()
			}
			s.tracker.state = stateCompleteUpdate

		case stateCompleteUpdate:
			cond := s.action.GetStatusConditions().Completed
			if err := UpdateStatus(r.Client, s.cr, string(cond), cond); err != nil {
				return controllerutils.NewRequeueWithShortDelay()
			}
			s.tracker.state = stateEnd
		}
	}
	return ctrl.Result{}
}

func loadHelmInfo(cr *moduleplatform.ModuleLifecycle) compspi.HelmInfo {
	helmInfo := compspi.HelmInfo{
		HelmRelease: cr.Spec.Installer.HelmRelease,
	}
	return helmInfo
}

func (r *Reconciler) getAction(action moduleplatform.ActionType) compspi.LifecycleActionHandler {
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

func procResult(res ctrl.Result, err error) ctrl.Result {
	if vzctrl.ShouldRequeue(res) {
		if res.RequeueAfter == 0 {
			return controllerutils.NewRequeueWithShortDelay()
		}
		return res
	}
	if err != nil {
		return controllerutils.NewRequeueWithShortDelay()
	}
	return res
}
