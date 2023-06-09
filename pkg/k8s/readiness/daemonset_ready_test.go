// Copyright (c) 2022, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package readiness

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	k8scheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestDaemonSetsReady(t *testing.T) {

	selector := &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "foo",
		},
	}
	enoughReplicas := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: selector,
		},
		Status: appsv1.DaemonSetStatus{
			DesiredNumberScheduled: 1,
			UpdatedNumberScheduled: 1,
		},
	}
	enoughReplicasMultiple := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: selector,
		},
		Status: appsv1.DaemonSetStatus{
			DesiredNumberScheduled: 2,
			UpdatedNumberScheduled: 2,
		},
	}
	unavalableNodes := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: selector,
		},
		Status: appsv1.DaemonSetStatus{
			DesiredNumberScheduled: 2,
			UpdatedNumberScheduled: 2,
			NumberUnavailable:      1,
		},
	}
	notEnoughReadyReplicas := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: selector,
		},
		Status: appsv1.DaemonSetStatus{
			DesiredNumberScheduled: 0,
			UpdatedNumberScheduled: 1,
		},
	}
	notEnoughUpdatedReplicas := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: selector,
		},
		Status: appsv1.DaemonSetStatus{
			DesiredNumberScheduled: 1,
			UpdatedNumberScheduled: 0,
		},
	}

	readyPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo-95d8c5d96-m6mbr",
			Namespace: "bar",
			Labels: map[string]string{
				controllerRevisionHashLabel: "95d8c5d96",
				"app":                       "foo",
			},
		},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Ready: true,
				},
			},
		},
	}
	notReadyContainerPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo-95d8c5d96-m6y76",
			Namespace: "bar",
			Labels: map[string]string{
				controllerRevisionHashLabel: "95d8c5d96",
				"app":                       "foo",
			},
		},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Ready: false,
				},
			},
		},
	}
	notReadyInitContainerPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo-95d8c5d96-m6mbr",
			Namespace: "bar",
			Labels: map[string]string{
				controllerRevisionHashLabel: "95d8c5d96",
				"app":                       "foo",
			},
		},
		Status: corev1.PodStatus{
			InitContainerStatuses: []corev1.ContainerStatus{
				{
					Ready: false,
				},
			},
		},
	}
	noPodHashLabel := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo-95d8c5d96-m6mbr",
			Namespace: "bar",
			Labels: map[string]string{
				"app": "foo",
			},
		},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Ready: true,
				},
			},
		},
	}

	controllerRevision := &appsv1.ControllerRevision{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo-95d8c5d96",
			Namespace: "bar",
		},
		Revision: 1,
	}

	namespacedName := []types.NamespacedName{
		{
			Name:      "foo",
			Namespace: "bar",
		},
	}

	var tests = []struct {
		name  string
		c     client.Client
		n     []types.NamespacedName
		ready bool
	}{
		{
			"should be ready when daemonset has enough replicas and pod is ready",
			fake.NewClientBuilder().WithScheme(k8scheme.Scheme).WithObjects(enoughReplicas, readyPod, controllerRevision).Build(),
			namespacedName,
			true,
		},
		{
			"should not be ready when daemonset has enough replicas and one pod of two pods is not ready",
			fake.NewClientBuilder().WithScheme(k8scheme.Scheme).WithObjects(enoughReplicasMultiple, notReadyContainerPod, readyPod, controllerRevision).Build(),
			namespacedName,
			false,
		},
		{
			"should be not ready when expected pods ready not reached",
			fake.NewClientBuilder().WithScheme(k8scheme.Scheme).WithObjects(enoughReplicasMultiple, readyPod, controllerRevision).Build(),
			namespacedName,
			false,
		},
		{
			"should be not ready when daemonset has enough replicas but pod init container pods is not ready",
			fake.NewClientBuilder().WithScheme(k8scheme.Scheme).WithObjects(enoughReplicas, notReadyInitContainerPod, controllerRevision).Build(),
			namespacedName,
			false,
		},
		{
			"should be not ready when daemonset has enough replicas but pod container pods is not ready",
			fake.NewClientBuilder().WithScheme(k8scheme.Scheme).WithObjects(enoughReplicas, notReadyContainerPod, controllerRevision).Build(),
			namespacedName,
			false,
		},
		{
			"should be not ready when daemonset has enough replicas but some nodes don't have pods ready",
			fake.NewClientBuilder().WithScheme(k8scheme.Scheme).WithObjects(unavalableNodes, readyPod, controllerRevision).Build(),
			namespacedName,
			false,
		},
		{
			"should be not ready when pod not found for daemonset",
			fake.NewClientBuilder().WithScheme(k8scheme.Scheme).WithObjects(enoughReplicas).Build(),
			namespacedName,
			false,
		},
		{
			"should be not ready when daemonset not found",
			fake.NewClientBuilder().WithScheme(k8scheme.Scheme).Build(),
			namespacedName,
			false,
		},
		{
			"should be not ready when controller-revision-hash label not found",
			fake.NewClientBuilder().WithScheme(k8scheme.Scheme).WithObjects(enoughReplicas, noPodHashLabel).Build(),
			namespacedName,
			false,
		},
		{
			"should be not ready when controllerrevision resource not found",
			fake.NewClientBuilder().WithScheme(k8scheme.Scheme).WithObjects(enoughReplicas, readyPod).Build(),
			namespacedName,
			false,
		},
		{
			"should be not ready when daemonset doesn't have enough ready replicas",
			fake.NewClientBuilder().WithScheme(k8scheme.Scheme).WithObjects(notEnoughReadyReplicas).Build(),
			namespacedName,
			false,
		},
		{
			"should be not ready when daemonset doesn't have enough updated replicas",
			fake.NewClientBuilder().WithScheme(k8scheme.Scheme).WithObjects(notEnoughUpdatedReplicas).Build(),
			namespacedName,
			false,
		},
		{
			"should be ready when no PodReadyCheck provided",
			fake.NewClientBuilder().WithScheme(k8scheme.Scheme).Build(),
			[]types.NamespacedName{},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.ready, DaemonSetsAreReady(vzlog.DefaultLogger(), tt.c, tt.n, ""))
		})
	}
}

// TestDaemonSetGenerationMismatch tests a daemonset generation ready status check
// GIVEN a call validate DaemonSetsAreReady
// WHEN the generation does not match the observed generation
// THEN false is returned
func TestDaemonSetGenerationMismatch(t *testing.T) {
	namespacedName := []types.NamespacedName{
		{
			Name:      "foo",
			Namespace: "bar",
		},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(k8scheme.Scheme).WithObjects(
		&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:  "bar",
				Name:       "foo",
				Generation: 4,
			},
			Status: appsv1.DaemonSetStatus{
				ObservedGeneration: 3,
			},
		}).Build()

	assert.False(t, DaemonSetsAreReady(vzlog.DefaultLogger(), fakeClient, namespacedName, ""))
}
