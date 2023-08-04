// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package module

import (
	"github.com/verrazzano/verrazzano-modules/pkg/controller/base/controllerspi"
	corev1 "k8s.io/api/core/v1"
	clipkg "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// GetDefaultWatchDescriptors returns the list of WatchDescriptors for objects being watched by the Module
// Always for secrets and configmaps since they may contain module configuration
func (r *Reconciler) GetDefaultWatchDescriptors() []controllerspi.WatchDescriptor {
	return []controllerspi.WatchDescriptor{
		{
			WatchedResourceKind: source.Kind{Type: &corev1.Secret{}},
			FuncShouldReconcile: r.ShouldSecretEventTriggerReconcile,
		},
		{
			WatchedResourceKind: source.Kind{Type: &corev1.ConfigMap{}},
			FuncShouldReconcile: r.ShouldConfigmapEventTriggerReconcile,
		},
	}
}

// ShouldSecretEventTriggerReconcile returns true if reconcile should be done in response to a Secret lifecycle event
func (r *Reconciler) ShouldSecretEventTriggerReconcile(obj clipkg.Object, event controllerspi.WatchEvent) bool {
	if event == controllerspi.Deleted {
		return false
	}
	secret := obj.(*corev1.Secret)
	return doesModuleOwnResource(secret.Labels, r.Name())
}

// ShouldConfigmapEventTriggerReconcile returns true if reconcile should be done in response to a Configmap lifecycle event
func (r *Reconciler) ShouldConfigmapEventTriggerReconcile(obj clipkg.Object, event controllerspi.WatchEvent) bool {
	if event == controllerspi.Deleted {
		return false
	}
	cm := obj.(*corev1.ConfigMap)
	return doesModuleOwnResource(cm.Labels, h.Name())
}

// doesModuleOwnResource returns true if the resource module owner label matches component
func doesModuleOwnResource(labels map[string]string, moduleName string) bool {
	if labels == nil {
		return false
	}
	owner, ok := labels[constants.VerrazzanoModuleOwnerLabel]
	if !ok {
		return false
	}
	// return true if this resource has the module owner label that matches this component.
	return owner == moduleName
}
