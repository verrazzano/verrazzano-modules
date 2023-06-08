// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package k8sutil

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	k8scheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	objects                  = "./testdata/objects"
	testdata                 = "./testdata"
	verrazzanoModuleOperator = "verrazzano-platform-operator"
	verrazzanoInstall        = "verrazzano-install"
)

// TestApplyD
// GIVEN valid objects and invalid
//
//	WHEN I call apply with changes
//	THEN the resulting object contains the updates as expected
func TestApplyD(t *testing.T) {
	var tests = []struct {
		name    string
		dir     string
		count   int
		isError bool
	}{
		{
			"should apply YAML files",
			objects,
			3,
			false,
		},
		{
			"should fail to apply non-existent directories",
			"blahblah",
			0,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := fake.NewClientBuilder().WithScheme(k8scheme.Scheme).Build()
			y := NewYAMLApplier(c, "")
			err := y.ApplyD(tt.dir)
			if tt.isError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.count, len(y.Objects()))
			}
		})
	}
}

// TestApplyF
// GIVEN a list of fields
//
//	WHEN I call apply with changes on fields
//	THEN the resulting object contains the updates as expected
func TestApplyF(t *testing.T) {
	var tests = []struct {
		name                                 string
		file                                 string
		count                                int
		isError                              bool
		expectedLastAppliedConfigAnnotations []string
	}{
		{
			"should apply file",
			objects + "/service.yaml",
			1,
			false,
			[]string{"{\"apiVersion\":\"v1\",\"kind\":\"Service\",\"metadata\":{\"annotations\":{},\"name\":\"my-service\",\"namespace\":\"test\"},\"spec\":{\"ports\":[{\"port\":80,\"protocol\":\"TCP\",\"targetPort\":9376}],\"selector\":{\"app\":\"MyApp\"}}}\n"},
		},
		{
			"should apply file with two objects",
			testdata + "/two_objects.yaml",
			2,
			false,
			[]string{"{\"apiVersion\":\"v1\",\"kind\":\"Service\",\"metadata\":{\"annotations\":{},\"name\":\"service1\",\"namespace\":\"test\"},\"spec\":{\"ports\":[{\"port\":80,\"protocol\":\"TCP\",\"targetPort\":9376}],\"selector\":{\"app\":\"MyApp\"}}}\n",
				"{\"apiVersion\":\"v1\",\"kind\":\"Service\",\"metadata\":{\"annotations\":{},\"name\":\"service2\",\"namespace\":\"test\"},\"spec\":{\"ports\":[{\"port\":80,\"protocol\":\"TCP\",\"targetPort\":9376}],\"selector\":{\"app\":\"MyApp\"}}}\n"},
		},
		{
			"should fail to apply files that are not YAML",
			"blahblah",
			0,
			true,
			nil,
		},
		{
			"should fail when file is not YAML",
			objects + "/not-yaml.txt",
			0,
			true,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := fake.NewClientBuilder().WithScheme(k8scheme.Scheme).Build()
			y := NewYAMLApplier(c, "test")
			err := y.ApplyF(tt.file)
			if tt.isError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.count, len(y.Objects()))

			for i, actualObj := range y.Objects() {
				actual := actualObj.GetAnnotations()[corev1.LastAppliedConfigAnnotation]
				expected := tt.expectedLastAppliedConfigAnnotations[i]
				if diff := cmp.Diff(actual, expected); diff != "" {
					t.Errorf("expected %v\n, got %v instead", expected, actual)
					t.Logf("Difference: %s", diff)
				}
			}
		})
	}
}

// TestApplyFNonSpec
// GIVEN a object that contains top level fields outside of spec
//
//	WHEN I call apply with changes non-spec fields
//	THEN the resulting object contains the updates
func TestApplyFNonSpec(t *testing.T) {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      verrazzanoModuleOperator,
			Namespace: verrazzanoInstall,
		},
		Secrets: []corev1.ObjectReference{
			{
				Name: "verrazzano-platform-operator-token",
			},
		},
	}
	c := fake.NewClientBuilder().WithScheme(newScheme()).WithObjects(sa).Build()
	y := NewYAMLApplier(c, "")
	err := y.ApplyF(testdata + "/sa_add_imagepullsecrets.yaml")
	assert.NoError(t, err)

	// Verify the resulting SA
	saUpdated := &corev1.ServiceAccount{}
	err = c.Get(context.TODO(), types.NamespacedName{Name: verrazzanoModuleOperator, Namespace: verrazzanoInstall}, saUpdated)
	assert.NoError(t, err)

	assert.NotEmpty(t, saUpdated.ImagePullSecrets)
	assert.Equal(t, 1, len(saUpdated.ImagePullSecrets))
	assert.Equal(t, "verrazzano-container-registry", saUpdated.ImagePullSecrets[0].Name)

	assert.Empty(t, saUpdated.Secrets)
	assert.Equal(t, 0, len(saUpdated.Secrets))
}

// TestApplyFMerge
// GIVEN a object that contains spec field
//
//	WHEN I call apply with additions to the spec field
//	THEN the resulting object contains the merged updates
func TestApplyFMerge(t *testing.T) {
	deadlineSeconds := int32(5)
	deployment := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      verrazzanoModuleOperator,
			Namespace: verrazzanoInstall,
		},
		Spec: appv1.DeploymentSpec{
			MinReadySeconds:         5,
			ProgressDeadlineSeconds: &deadlineSeconds,
		},
	}
	c := fake.NewClientBuilder().WithScheme(newScheme()).WithObjects(deployment).Build()
	y := NewYAMLApplier(c, "")
	err := y.ApplyF(testdata + "/deployment_merge.yaml")
	assert.NoError(t, err)

	// Verify the resulting Deployment
	depUpdated := &appv1.Deployment{}
	err = c.Get(context.TODO(), types.NamespacedName{Name: verrazzanoModuleOperator, Namespace: verrazzanoInstall}, depUpdated)
	assert.NoError(t, err)

	assert.Equal(t, int32(5), depUpdated.Spec.MinReadySeconds)
	assert.Equal(t, int32(5), *depUpdated.Spec.Replicas)
	assert.Equal(t, int32(10), *depUpdated.Spec.ProgressDeadlineSeconds)
}

// TestApplyFClusterRole
// GIVEN a ClusterRole object
//
//	WHEN I call apply with additions
//	THEN the resulting object contains the merged updates
func TestApplyFClusterRole(t *testing.T) {
	deadlineSeconds := int32(5)
	deployment := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      verrazzanoModuleOperator,
			Namespace: verrazzanoInstall,
		},
		Spec: appv1.DeploymentSpec{
			MinReadySeconds:         5,
			ProgressDeadlineSeconds: &deadlineSeconds,
		},
	}
	c := fake.NewClientBuilder().WithScheme(newScheme()).WithObjects(deployment).Build()
	y := NewYAMLApplier(c, "")
	err := y.ApplyF(testdata + "/clusterrole_create.yaml")
	assert.NoError(t, err)

	// Verify the ClusterRole that was created
	clusterRole := &rbacv1.ClusterRole{}
	err = c.Get(context.TODO(), types.NamespacedName{Name: "verrazzano-managed-cluster"}, clusterRole)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(clusterRole.Rules))
	rule := clusterRole.Rules[0]
	assert.Equal(t, "", rule.APIGroups[0])
	assert.Equal(t, "secrets", rule.Resources[0])
	assert.Equal(t, 3, len(rule.Verbs))

	// Update the ClusterRole
	err = y.ApplyF(testdata + "/clusterrole_update.yaml")
	assert.NoError(t, err)

	// Verify the ClusterRole that was updated
	clusterRoleUpdated := &rbacv1.ClusterRole{}
	err = c.Get(context.TODO(), types.NamespacedName{Name: "verrazzano-managed-cluster"}, clusterRoleUpdated)
	assert.NoError(t, err)
	rule = clusterRoleUpdated.Rules[0]
	assert.Equal(t, 4, len(rule.Verbs))

	// Verify all the expected verbs are there
	foundCount := 0
	for _, verb := range rule.Verbs {
		switch verb {
		case "get":
			foundCount++
		case "list":
			foundCount++
		case "watch":
			foundCount++
		case "update":
			foundCount++
		}
	}
	assert.Equal(t, 4, foundCount)
}

// TestApplyFT
// GIVEN a template files with valid and invalid info
//
//	WHEN I call file template spec
//	THEN the resulting object contains the expected configuration
func TestApplyFT(t *testing.T) {
	var tests = []struct {
		name                                 string
		file                                 string
		args                                 map[string]interface{}
		count                                int
		isError                              bool
		expectedLastAppliedConfigAnnotations []string
	}{
		{
			"should apply a template file",
			testdata + "/templated_service.yaml",
			map[string]interface{}{"namespace": "default"},
			1,
			false,
			[]string{"{\"apiVersion\":\"v1\",\"kind\":\"Service\",\"metadata\":{\"annotations\":{},\"name\":\"tmpl-service\",\"namespace\":\"default\"},\"spec\":{\"ports\":[{\"port\":80,\"protocol\":\"TCP\",\"targetPort\":9376}],\"selector\":{\"app\":\"MyApp\"}}}\n"},
		},
		{
			"should fail to apply when template is incomplete",
			testdata + "/templated_service.yaml",
			map[string]interface{}{},
			0,
			true,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := fake.NewClientBuilder().WithScheme(k8scheme.Scheme).Build()
			y := NewYAMLApplier(c, "")
			err := y.ApplyFT(tt.file, tt.args)
			if tt.isError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.count, len(y.Objects()))

			for i, actualObj := range y.Objects() {
				actual := actualObj.GetAnnotations()[corev1.LastAppliedConfigAnnotation]
				assert.NotEmpty(t, actual)
				expected := tt.expectedLastAppliedConfigAnnotations[i]
				if diff := cmp.Diff(actual, expected); diff != "" {
					t.Errorf("expected %v\n, got %v instead", expected, actual)
					t.Logf("Difference: %s", diff)
				}
			}
		})
	}
}

// TestApplyDT tests the ApplyDT function.
// GIVEN the directory of file templates
//
// WHEN ApplyDT is called
// THEN the result object should return error as expected
func TestApplyDT(t *testing.T) {
	var tests = []struct {
		name    string
		dir     string
		args    map[string]interface{}
		count   int
		isError bool
	}{
		// GIVEN a directory of template YAML files
		// WHEN the ApplyDT function is called with substitution key/value pairs
		// THEN the call succeeds and the resources are applied to the cluster
		{
			"should apply all template files in directory",
			testdata,
			map[string]interface{}{"namespace": "default"},
			7,
			false,
		},
		// GIVEN a directory of template YAML files
		// WHEN the ApplyDT function is called with no substitution key/value pairs
		// THEN the call fails
		{
			"should fail to apply when one or more templates are incomplete",
			testdata,
			map[string]interface{}{},
			4,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := fake.NewClientBuilder().WithScheme(k8scheme.Scheme).Build()
			y := NewYAMLApplier(c, "")
			err := y.ApplyDT(tt.dir, tt.args)
			if tt.isError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.count, len(y.Objects()))
		})
	}
}

// TestDeleteF tests the DeleteF function.
// GIVEN the spec file
//
// WHEN DeleteF is called
// THEN the function should return error as expected
func TestDeleteF(t *testing.T) {
	var tests = []struct {
		name    string
		file    string
		isError bool
	}{
		{
			"should delete valid file",
			testdata + "/two_objects.yaml",
			false,
		},
		{
			"should fail to delete invalid file",
			"blahblah",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := fake.NewClientBuilder().WithScheme(k8scheme.Scheme).Build()
			y := NewYAMLApplier(c, "")
			err := y.DeleteF(tt.file)
			if tt.isError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDeleteFD tests the DeleteF function.
// GIVEN the spec file template spec
//
// WHEN DeleteFT is called
// THEN the function should return error as expected
func TestDeleteFD(t *testing.T) {
	var tests = []struct {
		name    string
		file    string
		args    map[string]interface{}
		isError bool
	}{
		{
			"should apply a template file",
			testdata + "/templated_service.yaml",
			map[string]interface{}{"namespace": "default"},
			false,
		},
		{
			"should fail to apply when template is incomplete",
			testdata + "/templated_service.yaml",
			map[string]interface{}{},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := fake.NewClientBuilder().WithScheme(k8scheme.Scheme).Build()
			y := NewYAMLApplier(c, "")
			err := y.DeleteFT(tt.file, tt.args)
			if tt.isError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDeleteAll tests the ApplyD and DeleteAll function.
// GIVEN the list of valid objects
//
// WHEN ApplyD and DeleteAll is called
// THEN the function should return no errors
func TestDeleteAll(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(k8scheme.Scheme).Build()
	y := NewYAMLApplier(c, "")
	err := y.ApplyD(objects)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(y.Objects()))
	err = y.DeleteAll()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(y.Objects()))
}

func newScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = appv1.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)
	_ = k8scheme.AddToScheme(scheme)
	_ = rbacv1.AddToScheme(scheme)
	return scheme
}
