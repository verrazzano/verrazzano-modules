// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package common

import (
	"github.com/stretchr/testify/assert"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/handlerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakes "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

const (
	name     = "res1"
	ns       = "ns1"
	matchKey = "app.kubernetes.io/name"

	// pod label used to identify the controllerRevision resource for daemonsets and statefulsets
	controllerRevisionHashLabel = "controller-revision-hash"
)

// TestDeploymentReady tests the deployment readiness
// GIVEN a set of resources for a Helm release
// WHEN CheckWorkLoadsReady is called
// THEN ensure that correct readiness bool is returned.
func TestDeploymentReady(t *testing.T) {
	const stsRevision = "1"
	const stsRevisionNum = 1

	asserts := assert.New(t)
	tests := []struct {
		name          string
		releaseName   string
		replicas      int32
		readyReplicas int32
		expectedReady bool
		hemlLabelVal  string
	}{
		{
			name:          "test1",
			releaseName:   "rel1",
			replicas:      1,
			readyReplicas: 1,
			expectedReady: true,
		},
		{
			name:          "test2",
			releaseName:   "rel2",
			replicas:      2,
			readyReplicas: 1,
			expectedReady: false,
		},
		{
			name:          "test3-diff-helm",
			releaseName:   "rel2",
			replicas:      2,
			readyReplicas: 1,
			expectedReady: true,
			hemlLabelVal:  "wrongHelm",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			hemlLabelVal := test.releaseName
			if test.hemlLabelVal != "" {
				hemlLabelVal = test.hemlLabelVal
			}
			sts := v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name, Annotations: map[string]string{helmKey: hemlLabelVal}},
				Spec: v1.DeploymentSpec{
					Replicas: &test.replicas,
					Selector: &metav1.LabelSelector{
						MatchLabels:      map[string]string{matchKey: test.releaseName},
						MatchExpressions: nil,
					},
				},
				Status: v1.DeploymentStatus{
					ReadyReplicas: test.readyReplicas,
				},
			}
			pod := corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name,
					Labels: map[string]string{matchKey: test.releaseName,
						controllerRevisionHashLabel: stsRevision}},
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{{Ready: true}},
				},
			}

			cli := fakes.NewClientBuilder().WithScheme(newScheme()).WithObjects(&sts, &pod).Build()
			rctx := handlerspi.HandlerContext{
				Log:    vzlog.DefaultLogger(),
				Client: cli,
			}
			ready, err := CheckWorkLoadsReady(rctx, test.releaseName, ns)
			asserts.NoError(err)
			asserts.Equal(test.expectedReady, ready)
		})
	}
}

// TestStatefulSetReady tests the statefulset readiness
// GIVEN a set of resources for a Helm release
// WHEN CheckWorkLoadsReady is called
// THEN ensure that correct readiness bool is returned.
func TestStatefulSetReady(t *testing.T) {
	const stsRevision = "1"
	const stsRevisionNum = 1

	asserts := assert.New(t)
	tests := []struct {
		name          string
		releaseName   string
		replicas      int32
		readyReplicas int32
		expectedReady bool
	}{
		{
			name:          "test1",
			releaseName:   "rel1",
			replicas:      1,
			readyReplicas: 1,
			expectedReady: true,
		},
		{
			name:          "test2",
			releaseName:   "rel2",
			replicas:      2,
			readyReplicas: 1,
			expectedReady: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sts := v1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name, Annotations: map[string]string{helmKey: test.releaseName}},
				Spec: v1.StatefulSetSpec{
					Replicas: &test.replicas,
					Selector: &metav1.LabelSelector{
						MatchLabels:      map[string]string{matchKey: test.releaseName},
						MatchExpressions: nil,
					},
				},
				Status: v1.StatefulSetStatus{
					ReadyReplicas: test.readyReplicas,
				},
			}
			pod := corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name,
					Labels: map[string]string{matchKey: test.releaseName,
						controllerRevisionHashLabel: stsRevision}},
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{{Ready: true}},
				},
			}

			crev := appsv1.ControllerRevision{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      stsRevision,
					Namespace: ns,
				},
				Revision: stsRevisionNum,
			}

			cli := fakes.NewClientBuilder().WithScheme(newScheme()).WithObjects(&sts, &pod, &crev).Build()
			rctx := handlerspi.HandlerContext{
				Log:    vzlog.DefaultLogger(),
				Client: cli,
			}
			ready, err := CheckWorkLoadsReady(rctx, test.releaseName, ns)
			asserts.NoError(err)
			asserts.Equal(test.expectedReady, ready)
		})
	}
}

// TestDaemonsetReady tests the daemonset readiness
// GIVEN a set of resources for a Helm release
// WHEN CheckWorkLoadsReady is called
// THEN ensure that correct readiness bool is returned.
func TestDaemonsetReady(t *testing.T) {
	const stsRevision = "1"
	const stsRevisionNum = 1

	asserts := assert.New(t)
	tests := []struct {
		name                string
		releaseName         string
		unavailableReplicas int32
		expectedReady       bool
		hemlLabelVal        string
	}{
		{
			name:          "test1",
			releaseName:   "rel1",
			expectedReady: true,
		},
		{
			name:                "test2",
			releaseName:         "rel2",
			unavailableReplicas: 1,
			expectedReady:       false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			hemlLabelVal := test.releaseName
			if test.hemlLabelVal != "" {
				hemlLabelVal = test.hemlLabelVal
			}
			sts := v1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name, Annotations: map[string]string{helmKey: hemlLabelVal}},
				Spec: v1.DaemonSetSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels:      map[string]string{matchKey: test.releaseName},
						MatchExpressions: nil,
					},
				},
				Status: v1.DaemonSetStatus{
					NumberUnavailable: test.unavailableReplicas,
				},
			}
			pod := corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name,
					Labels: map[string]string{matchKey: test.releaseName,
						controllerRevisionHashLabel: stsRevision}},
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{{Ready: true}},
				},
			}

			cli := fakes.NewClientBuilder().WithScheme(newScheme()).WithObjects(&sts, &pod).Build()
			rctx := handlerspi.HandlerContext{
				Log:    vzlog.DefaultLogger(),
				Client: cli,
			}
			ready, err := CheckWorkLoadsReady(rctx, test.releaseName, ns)
			asserts.NoError(err)
			asserts.Equal(test.expectedReady, ready)
		})
	}
}

// newScheme creates a new scheme that includes this package's object to use for testing
func newScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = v1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	return scheme
}
