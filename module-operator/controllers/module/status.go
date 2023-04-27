// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package module

import (
	"context"
	"fmt"
	"time"

	modulesv1alpha1 "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// UpdateStatus configures the Module's status based on the passed in state and then updates the Module on the cluster
func UpdateStatus(client client.Client, module *modulesv1alpha1.Module, msg string, condition modulesv1alpha1.ModuleCondition) error {
	state := modulesv1alpha1.ModuleStateReconciling

	// Update the Module's State
	module.Status.State = modulesv1alpha1.ModuleStateType(state)

	// Append a new condition, if applicable
	AppendCondition(module, msg, condition)

	// Update the module lifecycle status
	return client.Status().Update(context.TODO(), module)
}

func NewCondition(message string, condition modulesv1alpha1.ModuleCondition) modulesv1alpha1.ModuleCondition {
	t := time.Now().UTC()
	return modulesv1alpha1.ModuleCondition{
		Type:    condition.Type,
		Message: message,
		Status:  corev1.ConditionTrue,
		LastTransitionTime: fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02dZ",
			t.Year(), t.Month(), t.Day(),
			t.Hour(), t.Minute(), t.Second()),
	}
}

func AppendCondition(module *modulesv1alpha1.Module, message string, condition modulesv1alpha1.ModuleCondition) {
	conditions := module.Status.Conditions
	newCondition := NewCondition(message, condition)
	var lastCondition modulesv1alpha1.ModuleCondition
	if len(conditions) > 0 {
		lastCondition = conditions[len(conditions)-1]
	}

	// Only update the conditions if there is a notable change between the last update
	if needsConditionUpdate(lastCondition, newCondition) {
		// Delete the oldest condition if at tracking limit
		if len(conditions) > modulesv1alpha1.ConditionArrayLimit {
			conditions = conditions[1:]
		}
		module.Status.Conditions = append(conditions, newCondition)
	}
}

// needsConditionUpdate checks if the condition needs an update
func needsConditionUpdate(last, new modulesv1alpha1.ModuleCondition) bool {
	return last.Type != new.Type && last.Message != new.Message
}
