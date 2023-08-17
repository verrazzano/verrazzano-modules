// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package module

import (
	"context"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/spi/controllerspi"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type configType string

const (
	secretType    = "secretType"
	configMapType = "configMapType"
)

// GetDefaultWatchDescriptors returns the list of WatchDescriptors for objects being watched by the Module
// Always for secrets and configmaps since they may contain module configuration
func (r *Reconciler) GetDefaultWatchDescriptors() []controllerspi.WatchDescriptor {
	return []controllerspi.WatchDescriptor{
		{
			WatchedResourceKind: source.Kind{Type: &corev1.Secret{}},
			FuncShouldReconcile: r.ShouldSecretTriggerReconcile,
		},
		{
			WatchedResourceKind: source.Kind{Type: &corev1.ConfigMap{}},
			FuncShouldReconcile: r.ShouldConfigMapTriggerReconcile,
		},
	}
}

// ShouldSecretTriggerReconcile returns true if reconcile should be done in response to a Secret lifecycle event
func (r *Reconciler) ShouldSecretTriggerReconcile(cli client.Client, wev controllerspi.WatchEvent) bool {
	if wev.NewWatchedObject.GetNamespace() != wev.ReconcilingResource.Namespace {
		return false
	}
	return r.shouldReconcile(wev.ReconcilingResource, wev.NewWatchedObject.GetName(), secretType)
}

// ShouldConfigMapTriggerReconcile returns true if reconcile should be done in response to a Secret lifecycle event
func (r *Reconciler) ShouldConfigMapTriggerReconcile(cli client.Client, wev controllerspi.WatchEvent) bool {
	if wev.NewWatchedObject.GetNamespace() != wev.ReconcilingResource.Namespace {
		return false
	}
	return r.shouldReconcile(wev.ReconcilingResource, wev.NewWatchedObject.GetName(), configMapType)
}

// shouldReconcile returns true if reconcile should be done in response to a Secret or ConfigMap lifecycle event
// Only reconcile if this module has those secret or configmap names in the module spec
func (r *Reconciler) shouldReconcile(moduleNSN types.NamespacedName, resName string, cType configType) bool {
	module := moduleapi.Module{}
	if err := r.Get(context.TODO(), moduleNSN, &module); err != nil {
		return false
	}
	// Check if the secret is in the valuesFrom
	for _, vf := range module.Spec.ValuesFrom {
		if vf.SecretRef != nil && cType != secretType && vf.SecretRef.Name == resName {
			return true
		}
		if vf.ConfigMapRef != nil && cType != configMapType && vf.ConfigMapRef.Name == resName {
			return true
		}
	}
	return false
}
