// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package lifecycle

import (
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/spi"
	compspi "github.com/verrazzano/verrazzano-modules/common/helm_component/spi"
	"github.com/verrazzano/verrazzano-modules/common/k8s"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"

	modplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	modulesv1alpha1 "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"

	vzctrl "github.com/verrazzano/verrazzano-modules/module-operator/pkg/controller"
	vzspi "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
)

// componentState identifies the state of a component during action
type componentState string

const (
	// stateInit is the state when a component is initialized
	stateInit componentState = "componentStateInit"

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
	cr        *modplatform.ModuleLifecycle
	tracker   *stateTracker
	chartInfo *compspi.HelmInfo
	action    compspi.LifecycleAction
}

// Reconcile updates the Certificate
func (r Reconciler) Reconcile(spictx spi.ReconcileContext, u *unstructured.Unstructured) (ctrl.Result, error) {
	cr := &modplatform.ModuleLifecycle{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, cr); err != nil {
		return ctrl.Result{}, err
	}

	ctx, err := vzspi.NewMinimalContext(r.Client, spictx.Log)
	if err != nil {
		return newRequeueWithDelay(), err
	}

	nsn := k8s.GetNamespacedName(cr.ObjectMeta)
	if cr.Generation == cr.Status.ObservedGeneration {
		spictx.Log.Debugf("Skipping reconcile for %v, observed generation has not change", nsn)
		return newRequeueWithDelay(), err
	}

	helmInfo := loadHelmInfo(cr)
	tracker := getTracker(cr.ObjectMeta, stateInit)

	smc := stateMachineContext{
		cr:        cr,
		tracker:   tracker,
		chartInfo: &helmInfo,
		action:    r.getAction("install"),
	}

	res, err := r.doStateMachine(ctx, smc)
	if err != nil {
		return newRequeueWithDelay(), err
	}
	if vzctrl.ShouldRequeue(res) {
		return res, nil
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) doStateMachine(spiCtx vzspi.ComponentContext, s stateMachineContext) (ctrl.Result, error) {
	compContext := spiCtx.Init("component").Operation("install") // TODO don't hard code this

	for s.tracker.state != stateEnd {
		switch s.tracker.state {
		case stateInit:
			res, err := s.action.Init(compContext, s.chartInfo)
			if err != nil {
				return ctrl.Result{}, err
			}
			if vzctrl.ShouldRequeue(res) {
				return res, nil
			}
			s.tracker.state = stateStartPreActionUpdate

		case stateStartPreActionUpdate:
			if err := UpdateStatus(r.Client, s.cr, string(modulesv1alpha1.CondPreInstall), modulesv1alpha1.CondPreInstall); err != nil {
				return ctrl.Result{}, err
			}
			s.tracker.state = statePreAction

		case statePreAction:
			res, err := s.action.PreAction(compContext)
			if err != nil {
				return ctrl.Result{}, err
			}
			if vzctrl.ShouldRequeue(res) {
				return res, nil
			}
			s.tracker.state = statePreActionWaitDone

		case statePreActionWaitDone:
			done, res, err := s.action.IsPreActionDone(compContext)
			if err != nil {
				return ctrl.Result{}, err
			}
			if vzctrl.ShouldRequeue(res) {
				return res, nil
			}
			if done {
				return ctrl.Result{}, nil
			}
			s.tracker.state = stateStartActionUpdate

		case stateStartActionUpdate:
			if err := UpdateStatus(r.Client, s.cr, string(modulesv1alpha1.CondInstallStarted), modulesv1alpha1.CondInstallStarted); err != nil {
				return ctrl.Result{}, err
			}
			s.tracker.state = stateAction

		case stateAction:
			res, err := s.action.DoAction(compContext)
			if err != nil {
				return ctrl.Result{}, err
			}
			if vzctrl.ShouldRequeue(res) {
				return res, nil
			}
			s.tracker.state = stateActionWaitDone

		case stateActionWaitDone:
			done, res, err := s.action.IsActionDone(compContext)
			if err != nil {
				return ctrl.Result{}, err
			}
			if vzctrl.ShouldRequeue(res) {
				return res, nil
			}
			if done {
				return ctrl.Result{}, nil
			}
			s.tracker.state = stateStartActionUpdate

		case statePostAction:
			res, err := s.action.PostAction(compContext)
			if err != nil {
				return ctrl.Result{}, err
			}
			if vzctrl.ShouldRequeue(res) {
				return res, nil
			}
			s.tracker.state = statePostActionWaitDone

		case statePostActionWaitDone:
			done, res, err := s.action.IsPreActionDone(compContext)
			if err != nil {
				return ctrl.Result{}, err
			}
			if vzctrl.ShouldRequeue(res) {
				return res, nil
			}
			if done {
				return ctrl.Result{}, nil
			}
			s.tracker.state = stateCompleteUpdate

		case stateCompleteUpdate:
			if err := UpdateStatus(r.Client, s.cr, string(modulesv1alpha1.CondUpgradeComplete), modulesv1alpha1.CondUpgradeComplete); err != nil {
				return ctrl.Result{}, err
			}
			s.tracker.state = stateEnd
		}
	}
	return ctrl.Result{}, nil
}

func newRequeueWithDelay() ctrl.Result {
	return vzctrl.NewRequeueWithDelay(3, 10, time.Second)
}

func loadHelmInfo(cr *modplatform.ModuleLifecycle) compspi.HelmInfo {
	helmInfo := compspi.HelmInfo{
		HelmRelease: cr.Spec.Installer.HelmRelease,
	}
	return helmInfo
}

func (r *Reconciler) getAction(action string) compspi.LifecycleAction {
	return r.comp.InstallAction
}
