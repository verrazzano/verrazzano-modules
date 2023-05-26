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

// TestReady tests the workload readiness
// GIVEN a set of resources for a Helm release
// WHEN CheckWorkLoadsReady is called
// THEN ensure that correct readiness bool is returned.
func TestReady(t *testing.T) {
	const stsRevision = "foo-95d8c5d96"

	asserts := assert.New(t)
	tests := []struct {
		name        string
		releaseName string
		namespace   string
		*v1.StatefulSet
		expectedReady bool
	}{
		{
			name:          "test1",
			releaseName:   "rel1",
			expectedReady: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sts := v1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name, Annotations: map[string]string{helmKey: test.releaseName}},
				Spec: v1.StatefulSetSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels:      map[string]string{matchKey: test.name},
						MatchExpressions: nil,
					},
				},
				Status: v1.StatefulSetStatus{
					ReadyReplicas:   1,
					UpdatedReplicas: 1,
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
					Namespace: test.Namespace,
				},
			}

			cli := fakes.NewClientBuilder().WithScheme(newScheme()).WithObjects(&sts, &pod, &crev).Build()
			rctx := handlerspi.HandlerContext{
				Log:    vzlog.DefaultLogger(),
				Client: cli,
			}
			ready, err := CheckWorkLoadsReady(rctx, test.releaseName, test.namespace)
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
