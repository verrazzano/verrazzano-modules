// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package reconciler

import (
	"context"
	"fmt"
	"time"

	modulesv1alpha1 "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// UpdateStatus configures the Module's status based on the passed in state and then updates the Module on the cluster
func UpdateStatus(client client.Client, mlc *modulesv1alpha1.ModuleLifecycle, msg string, condition modulesv1alpha1.LifecycleCondition) error {
	state := modulesv1alpha1.LifecycleState(condition)
	// Update the Module's State
	mlc.SetState(state)
	// Append a new condition, if applicable
	AppendCondition(mlc, msg, condition)

	// Update the module lifecycle status
	return client.Status().Update(context.TODO(), mlc)
}

func needsReconcile(mlc *modulesv1alpha1.ModuleLifecycle) bool {
	return mlc.Status.ObservedGeneration != mlc.Generation
}

func NewCondition(message string, condition modulesv1alpha1.LifecycleCondition) modulesv1alpha1.ModuleLifecycleCondition {
	t := time.Now().UTC()
	return modulesv1alpha1.ModuleLifecycleCondition{
		Type:    condition,
		Message: message,
		Status:  corev1.ConditionTrue,
		LastTransitionTime: fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02dZ",
			t.Year(), t.Month(), t.Day(),
			t.Hour(), t.Minute(), t.Second()),
	}
}

func AppendCondition(module *modulesv1alpha1.ModuleLifecycle, message string, condition modulesv1alpha1.LifecycleCondition) {
	conditions := module.Status.Conditions
	newCondition := NewCondition(message, condition)
	var lastCondition modulesv1alpha1.ModuleLifecycleCondition
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
func needsConditionUpdate(last, new modulesv1alpha1.ModuleLifecycleCondition) bool {
	return last.Type != new.Type && last.Message != new.Message
}
