// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package module

import (
	"context"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/status"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/statemachine"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/base/controllerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/handlerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/result"
	"github.com/verrazzano/verrazzano-modules/pkg/semver"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"strings"
	"time"

	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
)

var funcExecuteStateMachine = defaultExecuteStateMachine
var funcLoadHelmInfo = loadHelmInfo
var funcIsUpgradeNeeded = IsUpgradeNeeded
var ignoreHelmInfo bool

// IgnoreHelmInfo allows the module to ignore loading Helm info.  This is used for VPO integration.
func IgnoreHelmInfo() {
	ignoreHelmInfo = true
}

// Reconcile reconciles the Module CR
func (r Reconciler) Reconcile(spictx controllerspi.ReconcileContext, u *unstructured.Unstructured) result.Result {
	cr := &moduleapi.Module{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, cr); err != nil {
		spictx.Log.ErrorfThrottled(err.Error())
		// This is a fatal error, don't requeue
		return result.NewResult()
	}

	handlerCtx := handlerspi.HandlerContext{Client: r.Client, Log: spictx.Log}

	if cr.Generation == cr.Status.LastSuccessfulGeneration {
		return r.checkIfRequeueNeededWhenGenerationsMatch(cr)
	}

	handler, res := r.getActionHandler(handlerCtx, cr)
	if res.ShouldRequeue() {
		return res
	}
	if handler == nil {
		return result.NewResultShortRequeueDelay()
	}

	return r.reconcileAction(spictx, cr, handler)
}

// reconcileAction reconciles the Module CR for a particular action
func (r Reconciler) reconcileAction(spictx controllerspi.ReconcileContext, cr *moduleapi.Module, handler handlerspi.StateMachineHandler) result.Result {
	if cr.Spec.TargetNamespace == "" {
		cr.Spec.TargetNamespace = cr.Namespace
	}

	var err error

	// Temp hack to support VPO integration.  There is no Helm release for VPO module, but status update depends on it
	// so for now use module name/ns.
	helmInfo := handlerspi.HelmInfo{
		HelmRelease: &handlerspi.HelmRelease{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
	}
	if !ignoreHelmInfo {
		// Load the helm information needed by the handler
		helmInfo, err = funcLoadHelmInfo(cr)
		if err != nil {
			if strings.Contains(err.Error(), "FileNotFound") {
				spictx.Log.Errorf("Failed loading file information: %v", err)
				return result.NewResultRequeueDelay(10, 15, time.Second)
			}
			err := spictx.Log.ErrorfNewErr("Failed loading Helm info for %s/%s: %v", cr.Namespace, cr.Name, err)
			return result.NewResultShortRequeueDelayIfError(err)
		}
	}

	// Initialize the handler context
	ctx := handlerspi.HandlerContext{Client: r.Client, Log: spictx.Log, CR: cr, HelmInfo: helmInfo}

	// Execute the state machine
	sm := statemachine.StateMachine{
		Handler: handler,
		CR:      cr,
	}
	return funcExecuteStateMachine(ctx, sm)
}

// getActionHandler must return one of the Module action handlers.
func (r *Reconciler) getActionHandler(handlerCtx handlerspi.HandlerContext, cr *moduleapi.Module) (handlerspi.StateMachineHandler, result.Result) {
	if !status.IsInstalled(cr) {
		return r.ModuleHandlerInfo.InstallActionHandler, result.NewResult()
	}

	// return UpgradeAction only when the desired version is different from current
	upgradeNeeded, err := funcIsUpgradeNeeded(cr.Spec.Version, cr.Status.LastSuccessfulVersion)
	if err != nil {
		handlerCtx.Log.ErrorfThrottled("Failed checking if upgrade needed for Module %s/%s failed with error: %v\n", cr.Namespace, cr.Name, err)
		return nil, result.NewResultShortRequeueDelay()
	}
	if upgradeNeeded {
		return r.ModuleHandlerInfo.UpgradeActionHandler, result.NewResult()
	}
	return r.ModuleHandlerInfo.UpdateActionHandler, result.NewResult()

}

// IsUpgradeNeeded returns true if upgrade is needed
func IsUpgradeNeeded(desiredVersion, installedVersion string) (bool, error) {
	if len(desiredVersion) == 0 {
		return false, nil
	}
	desiredSemver, err := semver.NewSemVersion(desiredVersion)
	if err != nil {
		return false, err
	}
	installedSemver, err := semver.NewSemVersion(installedVersion)
	if err != nil {
		return false, err
	}
	return installedSemver.IsLessThan(desiredSemver), nil
}

func defaultExecuteStateMachine(ctx handlerspi.HandlerContext, sm statemachine.StateMachine) result.Result {
	return sm.Execute(ctx)
}

// checkIfRequeueNeededWhenGenerationsMatch determines if reconcile should be done
// when the cr.Generation matches the status generation, which means a previous
// reconcile successfully completed and updated the status generation.
// However, even if the reconciliation (e.g. install) finishes,
// reconcile might still get called a few times because controller-runtime can have
// CR updates in its cache. Also, a watched resource may have triggered an event causing
// reconcile to be called.  If the code was to continue to reconcile when it was really done,
// then the update action would occur and the Module condition would have update reasons
// instead of install reasons (e.g. InstallComplete).
//
// Therefore, we only re-reconcile if a watch triggered reconcile because
// something changed (the watched resource).  Determine if we need to reconcile
// based on the watch event timestamps.
func (r Reconciler) checkIfRequeueNeededWhenGenerationsMatch(module *moduleapi.Module) result.Result {
	watchEvent := r.BaseReconciler.GetLastWatchEvent(types.NamespacedName{Namespace: module.Namespace, Name: module.Name})
	if watchEvent == nil {
		// no watch events occurred
		return result.NewResult()
	}

	preInstallTime := statemachine.GetPreInstallTime(module)
	if preInstallTime != nil && watchEvent.EventTime.Before(*preInstallTime) {
		// watch event occurred before pre-install, so we can ignore it
		// since the pre-install and subsequent actions will use the latest resources
		return result.NewResult()
	}

	// Controller runtime generates Create event for all watched event on startup.
	// Ignore the Create event if the creation timestamp is older than 60 seconds, otherwise
	// every resource that uses watches will reconcile (like Module).
	// We can possibly remove this code when we optimize the module handlers. so they only call Helm
	// when needed by using a hash on the manifests, or something like that.
	if watchEvent.WatchEventType == controllerspi.Created {
		if watchEvent.WatchedResource.GetCreationTimestamp().Time.Add(time.Second * 60).Before(time.Now()) {
			return result.NewResult()
		}
	}

	// At this point, there was an event that happened after the last reconcile, so another reconcile needs to be done
	// Reset the last reconciled generation to zero so that we go through the normal reconcile loop
	// Also, reset the state machine tracker for this CR generation
	statemachine.DeleteTracker(module)
	module.Status.LastSuccessfulGeneration = 0
	err := r.Status().Update(context.TODO(), module)
	if err != nil {
		return result.NewResultShortRequeueDelayWithError(err)
	}
	return result.NewResultShortRequeueDelay()
}
