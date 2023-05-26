// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package common

import (
	"github.com/stretchr/testify/assert"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/handlerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	"path/filepath"
	fakes "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

const (
	name = "res1"
	ns   = "ns1"

	ccmFile = "ccm.yaml"
	//	calicoFile = "calico.yaml"
)

// TestCCM tests that the CCM workload readiness
// GIVEN a controller that implements the controllers spi interfaces
// WHEN Reconcile is called
// THEN ensure that the controller returns success and that the interface methods are all called
func TestCCM(t *testing.T) {
	asserts := assert.New(t)
	tests := []struct {
		name         string
		manifestFile string
		*v1.StatefulSet
		expectedReady bool
	}{
		{
			name:         "test1",
			manifestFile: ccmFile,
			StatefulSet: &v1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name},
				Status: v1.StatefulSetStatus{
					ReadyReplicas:   1,
					UpdatedReplicas: 1,
				},
			},
			expectedReady: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cli := fakes.NewClientBuilder().WithScheme(newScheme()).WithObjects(test.StatefulSet).Build()
			rctx := handlerspi.HandlerContext{
				Log:    vzlog.DefaultLogger(),
				Client: cli,
			}
			yam, err := os.ReadFile(filepath.Join("testdata", test.manifestFile))
			asserts.NoError(err)
			ready, err := checkWorkloadsReadyUsingManifest(rctx, "ccm", string(yam))
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
	return scheme
}
