// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package module

import (
	"context"
	"fmt"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/status"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/result"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/spi/controllerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/spi/handlerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/statemachine"
	"github.com/verrazzano/verrazzano-modules/pkg/semver"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"strings"
	"sync"
	"time"

	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
)

var funcExecuteStateMachine = defaultExecuteStateMachine
var funcLoadHelmInfo = loadHelmInfo
var funcIsUpgradeNeeded = IsUpgradeNeeded
var ignoreHelmInfo bool

var instanceMap = make(map[string]bool)
var maplock sync.Mutex

// IgnoreHelmInfo allows the module to ignore loading Helm info.  This is used for VPO integration.
func IgnoreHelmInfo() {
	ignoreHelmInfo = true
}

// Reconcile reconciles the Module CR
func (r Reconciler) Reconcile(spictx controllerspi.ReconcileContext, u *unstructured.Unstructured) result.Result {
	// Make sure controller doesn't call reconcile for the same instance concurrently
	if res := checkConcurrent(u); res.ShouldRequeue() {
		return res
	}
	defer deleteInstanceFromMap(u)

	cr := &moduleapi.Module{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, cr); err != nil {
		spictx.Log.ErrorfThrottled(err.Error())
		// This is a fatal error, don't requeue
		return result.NewResult()
	}

	// Initialize the handler context
	handlerCtx, res := r.initHandlerCtx(spictx, cr)
	if res.ShouldRequeue() {
		return res
	}

	// Check if this module was pre-installed by an external actor, like Verrazzano
	// This is needed for the non-module (Verrazzano component) to module upgrade case.
	if r.ModuleHandlerInfo.MigrationHandler != nil {
		res := r.ModuleHandlerInfo.MigrationHandler.UpdateStatusIfAlreadyInstalled(handlerCtx)
		if res.ShouldRequeue() {
			return res
		}
	}

	// If the CR has already been reconciled, see if a watch event occurred
	// that needs to be reconciled
	if cr.Generation == cr.Status.LastSuccessfulGeneration {
		return r.checkIfWatchEventRequiresReconcile(cr)
	}

	// Get the action handler
	handler, res := r.getActionHandler(handlerCtx, cr)
	if res.ShouldRequeue() {
		return res
	}
	if handler == nil {
		return result.NewResultShortRequeueDelay()
	}

	// Execute the state machine
	sm := statemachine.StateMachine{
		Handler: handler,
		CR:      cr,
	}
	res = funcExecuteStateMachine(handlerCtx, sm)
	if res.ShouldRequeue() {
		return res
	}

	// The normal state machine reconciliation has completed, meaning
	// the module has been installed, updated, or upgraded.
	// However, a watch may have occurred during the reconciliation,
	// maybe a watched component got installed.  Check if we need to re-reconcile.
	return r.checkIfWatchEventRequiresReconcile(cr)
}

// initHandlerCtx initializes the handler context
func (r Reconciler) initHandlerCtx(spictx controllerspi.ReconcileContext, cr *moduleapi.Module) (handlerspi.HandlerContext, result.Result) {
	// Default the target namespace to the cr namespace
	if cr.Spec.TargetNamespace == "" {
		cr.Spec.TargetNamespace = cr.Namespace
	}

	// Needed to support VPO integration.  There is no Helm release for VPO module, but status update depends on it
	// so for now use module name/ns.
	helmInfo := handlerspi.HelmInfo{
		HelmRelease: &handlerspi.HelmRelease{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
	}
	if !ignoreHelmInfo {
		// Load the helm information needed by the handler
		var err error
		helmInfo, err = funcLoadHelmInfo(cr)
		if err != nil {
			if strings.Contains(err.Error(), "FileNotFound") {
				spictx.Log.Errorf("Failed loading file information: %v", err)
				return handlerspi.HandlerContext{}, result.NewResultRequeueDelay(10, 15, time.Second)
			}
			err := spictx.Log.ErrorfNewErr("Failed loading Helm info for %s/%s: %v", cr.Namespace, cr.Name, err)
			return handlerspi.HandlerContext{}, result.NewResultShortRequeueDelayIfError(err)
		}
	}

	// Initialize the handler context
	handlerCtx := handlerspi.HandlerContext{Client: r.Client, Log: spictx.Log, CR: cr, HelmInfo: helmInfo}
	return handlerCtx, result.NewResult()
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

// checkIfWatchEventRequiresReconcile determines if reconcile should be done
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
func (r Reconciler) checkIfWatchEventRequiresReconcile(module *moduleapi.Module) result.Result {
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
		if watchEvent.NewWatchedObject.GetCreationTimestamp().Time.Add(time.Second * 60).Before(time.Now()) {
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

// checkConcurrent will prevent controller-runtime from calling reconcile on the same instance
// concurrently.
func checkConcurrent(u *unstructured.Unstructured) result.Result {
	// Make sure controller doesn't call reconcile for the same instance re-entrant
	maplock.Lock()
	defer func() { maplock.Unlock() }()

	nsn := getNsn(u)
	if _, ok := instanceMap[nsn]; ok {
		// Wait a few seconds to requeue
		return result.NewResultRequeueDelay(2, 3, time.Second)
	}

	instanceMap[nsn] = true
	return result.NewResult()
}

// deleteInstanceFromMap deletes the CR name from the instance map
func deleteInstanceFromMap(u *unstructured.Unstructured) {
	maplock.Lock()
	defer func() { maplock.Unlock() }()
	delete(instanceMap, getNsn(u))
}

func getNsn(u *unstructured.Unstructured) string {
	return fmt.Sprintf("%s-%s", u.GetNamespace(), u.GetName())
}
