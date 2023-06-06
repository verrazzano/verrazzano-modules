// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package common

import (
	"context"
	"fmt"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/handlerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/util"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

// UpdateReadyConditionSucceeded updates the Ready condition when the module has succeeded
func (h BaseHandler) UpdateReadyConditionSucceeded(ctx handlerspi.HandlerContext, reason moduleapi.ModuleConditionReason) (ctrl.Result, error) {
	h.ModuleCR.Status.LastSuccessfulVersion = h.ModuleCR.Spec.Version

	msgTemplate := readyConditionMessages[reason]
	msg := fmt.Sprintf(msgTemplate, h.ModuleCR.Spec.ModuleName, h.HelmRelease.Namespace, h.HelmRelease.Name)
	return h.updateReadyCondition(ctx, reason, corev1.ConditionTrue, msg)
}

// UpdateReadyConditionReconciling updates the Ready condition when the module is reconciling
func (h BaseHandler) UpdateReadyConditionReconciling(ctx handlerspi.HandlerContext, reason moduleapi.ModuleConditionReason) (ctrl.Result, error) {
	msgTemplate := readyConditionMessages[reason]
	msg := fmt.Sprintf(msgTemplate, h.ModuleCR.Spec.ModuleName, h.HelmRelease.Namespace, h.HelmRelease.Name)

	return h.updateReadyCondition(ctx, reason, corev1.ConditionFalse, msg)
}

// UpdateReadyConditionFailed updates the Ready condition when the module has failed
func (h BaseHandler) UpdateReadyConditionFailed(ctx handlerspi.HandlerContext, reason moduleapi.ModuleConditionReason, msgDetail string) (ctrl.Result, error) {
	msgTemplate := readyConditionMessages[reason]
	msg := fmt.Sprintf(msgTemplate, h.ModuleCR.Spec.ModuleName, h.HelmRelease.Namespace, h.HelmRelease.Name, msgDetail)

	return h.updateReadyCondition(ctx, reason, corev1.ConditionFalse, msg)
}

// updateReadyCondition updates the Ready condition
func (h BaseHandler) updateReadyCondition(ctx handlerspi.HandlerContext, reason moduleapi.ModuleConditionReason, status corev1.ConditionStatus, msg string) (ctrl.Result, error) {
	cond := moduleapi.ModuleCondition{
		Type:    moduleapi.ModuleConditionReady,
		Reason:  reason,
		Status:  status,
		Message: msg,
	}
	appendCondition(h.ModuleCR, cond)
	if err := ctx.Client.Status().Update(context.TODO(), h.ModuleCR); err != nil {
		return util.NewRequeueWithShortDelay(), nil
	}
	return ctrl.Result{}, nil
}

// appendCondition appends the condition to the list of conditions
func appendCondition(module *moduleapi.Module, cond moduleapi.ModuleCondition) {
	cond.LastTransitionTime = getTransitionTime()

	// Copy conditions that have a different type than the input condition into a new list
	var newConditions []moduleapi.ModuleCondition
	for i, existing := range module.Status.Conditions {
		if existing.Type != cond.Type {
			newConditions = append(newConditions, module.Status.Conditions[i])
		}
	}
	newConditions = append(newConditions, cond)
}

func getTransitionTime() string {
	t := time.Now().UTC()
	return fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02dZ",
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
}
