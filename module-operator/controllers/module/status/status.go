// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package status

import (
	"context"
	"fmt"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/handlerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

type StatusManager struct {
	ModuleName string
	*moduleapi.Module
	ReleaseNsn types.NamespacedName
}

// readyConditionMessages defines the condition messages for the Ready type condition
var readyConditionMessages = map[moduleapi.ModuleConditionReason]string{
	moduleapi.ReadyReasonInstallStarted:     "Started installing Module %s as Helm release %s/%s",
	moduleapi.ReadyReasonInstallSucceeded:   "Successfully installed Module %s as Helm release %s/%s",
	moduleapi.ReadyReasonInstallFailed:      "Failed installing Module %s as Helm release %s%s: %v",
	moduleapi.ReadyReasonUninstallStarted:   "Started uninstalling Module %s as Helm release %s/%s",
	moduleapi.ReadyReasonUninstallSucceeded: "Successfully uninstalled Module %s as Helm release %s/%s",
	moduleapi.ReadyReasonUninstallFailed:    "Failed uninstalling Module %s as Helm release %s/%s: %v",
	moduleapi.ReadyReasonUpdateStarted:      "Started updating Module %s as Helm release %s/%s",
	moduleapi.ReadyReasonUpdateSucceeded:    "Successfully updated Module %s as Helm release %s/%s",
	moduleapi.ReadyReasonUpdateFailed:       "Failed updating Module %s as Helm release %s/%s: %v",
	moduleapi.ReadyReasonUpgradeStarted:     "Started upgrading Module %s as Helm release %s/%s",
	moduleapi.ReadyReasonUpgradeSucceeded:   "Successfully upgraded Module %s as Helm release %s/%s",
	moduleapi.ReadyReasonUpgradeFailed:      "Failed upgrading Module %s as Helm release %s/%s: %v",
}

// UpdateReadyConditionSucceeded updates the Ready condition when the module has succeeded
func (s StatusManager) UpdateReadyConditionSucceeded(ctx handlerspi.HandlerContext, reason moduleapi.ModuleConditionReason, version string) (ctrl.Result, error) {
	s.Module.Status.LastSuccessfulVersion = version

	msgTemplate := readyConditionMessages[reason]
	msg := fmt.Sprintf(msgTemplate, s.ModuleName, s.ReleaseNsn.Namespace, s.ReleaseNsn.Name)
	return s.updateReadyCondition(ctx, reason, corev1.ConditionTrue, msg)
}

// UpdateReadyConditionReconciling updates the Ready condition when the module is reconciling
func (s StatusManager) UpdateReadyConditionReconciling(ctx handlerspi.HandlerContext, reason moduleapi.ModuleConditionReason) (ctrl.Result, error) {
	msgTemplate := readyConditionMessages[reason]
	msg := fmt.Sprintf(msgTemplate, s.ModuleName, s.ReleaseNsn.Namespace, s.ReleaseNsn.Name)

	return s.updateReadyCondition(ctx, reason, corev1.ConditionFalse, msg)
}

// UpdateReadyConditionFailed updates the Ready condition when the module has failed
func (s StatusManager) UpdateReadyConditionFailed(ctx handlerspi.HandlerContext, reason moduleapi.ModuleConditionReason, msgDetail string) (ctrl.Result, error) {
	msgTemplate := readyConditionMessages[reason]
	msg := fmt.Sprintf(msgTemplate, s.ModuleName, s.ReleaseNsn.Namespace, s.ReleaseNsn.Name, msgDetail)

	return s.updateReadyCondition(ctx, reason, corev1.ConditionFalse, msg)
}

// updateReadyCondition updates the Ready condition
func (s StatusManager) updateReadyCondition(ctx handlerspi.HandlerContext, reason moduleapi.ModuleConditionReason, status corev1.ConditionStatus, msg string) (ctrl.Result, error) {
	cond := moduleapi.ModuleCondition{
		Type:    moduleapi.ModuleConditionReady,
		Reason:  reason,
		Status:  status,
		Message: msg,
	}
	s.appendCondition(cond)
	if err := ctx.Client.Status().Update(context.TODO(), s.Module); err != nil {
		return util.NewRequeueWithShortDelay(), nil
	}
	return ctrl.Result{}, nil
}

// appendCondition appends the condition to the list of conditions
func (s StatusManager) appendCondition(cond moduleapi.ModuleCondition) {
	cond.LastTransitionTime = getTransitionTime()

	// Copy conditions that have a different type than the input condition into a new list
	var newConditions []moduleapi.ModuleCondition
	for i, existing := range s.Module.Status.Conditions {
		if existing.Type != cond.Type {
			newConditions = append(newConditions, s.Module.Status.Conditions[i])
		}
	}
	newConditions = append(newConditions, cond)
}

// IsInstalled checks if the modules is installed
func IsInstalled(cr *moduleapi.Module) bool {
	cond := GetReadyCondition(cr)
	if cond == nil {
		return false
	}

	// If the reason is not install started or failed, then assume installed.
	switch cond.Reason {
	case moduleapi.ReadyReasonInstallStarted:
	case moduleapi.ReadyReasonInstallFailed:
		return false
	default:
		return true
	}

	return false
}

// GetReadyCondition gets the Ready condition type
func GetReadyCondition(cr *moduleapi.Module) *moduleapi.ModuleCondition {
	for i, cond := range cr.Status.Conditions {
		if cond.Type == moduleapi.ModuleConditionReady {
			return &cr.Status.Conditions[i]
		}
	}
	return nil
}

func getTransitionTime() string {
	t := time.Now().UTC()
	return fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02dZ",
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
}
