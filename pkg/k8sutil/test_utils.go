// Copyright (c) 2022, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package k8sutil

import (
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	dynfake "k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	appsv1Cli "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1Cli "k8s.io/client-go/kubernetes/typed/core/v1"
)

// MockGetCoreV1 mocks GetCoreV1Client function
func MockGetCoreV1(objects ...runtime.Object) func(_ ...vzlog.VerrazzanoLogger) (corev1Cli.CoreV1Interface, error) {
	return func(_ ...vzlog.VerrazzanoLogger) (corev1Cli.CoreV1Interface, error) {
		return k8sfake.NewSimpleClientset(objects...).CoreV1(), nil
	}
}

// MockGetAppsV1 mocks GetAppsV1Client function
func MockGetAppsV1(objects ...runtime.Object) func(_ ...vzlog.VerrazzanoLogger) (appsv1Cli.AppsV1Interface, error) {
	return func(_ ...vzlog.VerrazzanoLogger) (appsv1Cli.AppsV1Interface, error) {
		return k8sfake.NewSimpleClientset(objects...).AppsV1(), nil
	}
}

// MockDynamicClient mocks GetDynamicClient function
func MockDynamicClient(objects ...runtime.Object) func() (dynamic.Interface, error) {
	return func() (dynamic.Interface, error) {
		return dynfake.NewSimpleDynamicClient(runtime.NewScheme(), objects...), nil
	}
}

func MkSvc(ns, name string) *corev1.Service {
	svc := &corev1.Service{}
	svc.Namespace = ns
	svc.Name = name
	return svc
}

func MkDep(ns, name string) *appsv1.Deployment {
	dep := &appsv1.Deployment{}
	dep.Namespace = ns
	dep.Name = name
	return dep
}
