// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package readiness

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	k8scheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// TestEnsurePodsAreReady tests the EnsurePodsAreReady function
// GIVEN a set of pods with their ready status and expected pods count
//
// WHEN EnsurePodsReady function is called on the inputs
// THEN the function should return pods ready count and isAllPodsReady as expected
func TestEnsurePodsAreReady(t *testing.T) {
	log := vzlog.DefaultLogger()
	tests := []struct {
		testName         string
		podsList         []corev1.Pod
		expectedPodsSize int32
		shouldPodReady   bool
	}{
		{
			testName:         "No pods",
			podsList:         []corev1.Pod{},
			expectedPodsSize: 0,
			shouldPodReady:   true,
		},
		{
			testName: "All pods are ready",
			podsList: []corev1.Pod{
				getMockPod(true, "pod1"),
				getMockPod(true, "pod2"),
			},
			expectedPodsSize: 2,
			shouldPodReady:   true,
		},
		{
			testName: "Only some pods are ready",
			podsList: []corev1.Pod{
				getMockPod(true, "pod3"),
				getMockPod(false, "pod4"),
			},
			expectedPodsSize: 1,
			shouldPodReady:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			numOfPodsReady, isAllPodsReady := EnsurePodsAreReady(log, tt.podsList, tt.expectedPodsSize, "prefix")
			assert.Equal(t, tt.expectedPodsSize, numOfPodsReady)
			assert.Equal(t, tt.shouldPodReady, isAllPodsReady)
		})
	}
}

func getMockPod(ready bool, podName string) corev1.Pod {
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
		},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Ready: ready,
				},
			},
		},
	}
}

// TestGetPodsList tests the GetPodsList function
// GIVEN a kubernetes client, with namespace and labels
//
// WHEN the GetPodsList function is called with the test objects
// THEN the function should return pods list that match with the given labels and namespace
func TestGetPodsList(t *testing.T) {
	mockpodWithOutLabel := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "testns",
		},
	}
	mockpodWithLabel := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod-with-label",
			Namespace: "testns",
			Labels: map[string]string{
				"app": "myapp",
			},
		},
	}
	log := vzlog.DefaultLogger()
	var tests = []struct {
		name          string
		c             client.Client
		namespace     types.NamespacedName
		labelSelector *metav1.LabelSelector
		expectedSize  int
	}{
		{
			"no pods present",
			fake.NewClientBuilder().WithScheme(k8scheme.Scheme).Build(),
			types.NamespacedName{
				Namespace: "testns",
				Name:      "testns",
			},
			&metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "myapp",
				},
				MatchExpressions: nil,
			},
			0,
		},
		{
			"pods present with given namespace",
			fake.NewClientBuilder().WithScheme(k8scheme.Scheme).WithObjects(mockpodWithOutLabel).WithObjects(mockpodWithLabel).Build(),
			types.NamespacedName{
				Namespace: "testns",
				Name:      "testns",
			},
			&metav1.LabelSelector{
				MatchLabels:      nil,
				MatchExpressions: nil,
			},
			2,
		},
		{
			"pods present with given namespace and label",
			fake.NewClientBuilder().WithScheme(k8scheme.Scheme).WithObjects(mockpodWithLabel).WithObjects(mockpodWithOutLabel).Build(),
			types.NamespacedName{
				Namespace: "testns",
				Name:      "testns",
			},
			&metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "myapp",
				},
				MatchExpressions: nil,
			},
			1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			podsList := GetPodsList(log, tt.c, tt.namespace, tt.labelSelector)
			assert.Equal(t, tt.expectedSize, len(podsList.Items))
		})
	}
}
