// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package helm_module_lifecycle_test

import (
	"context"
	"testing"

	"github.com/onsi/gomega"
	api "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/tests/common"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// TestOverrideOptionsLifecycle tests the module lifecycle of a module CR with different overrides.
// GIVEN an installation of module-operator in a cluster
// WHEN helm module version 0.1.0 is installed in a random namespace with overrides specified as inline, in a secret and configmap
// THEN the helm release for helm module is created in that namespace
// AND the module status eventually changes to ready
// AND the helm release values match to that of the overrides.
//
// GIVEN an installation of module-operator in a cluster
// AND helm module version 0.1.0 installed in a random namespace with overrides specified as inline, in a secret and configmap
// WHEN overrides are updated for the installed module in that namespace
// THEN the values for helm release for helm module is updated in that namespace
// AND the module status eventually changes to ready
// AND the helm release values match to that of the updated overrides.
//
// GIVEN an installation of module-operator in a cluster
// AND helm module version 0.1.0 installed in a random namespace with overrides specified as inline, in a secret and configmap
// WHEN helm module version is updated to 0.1.1 alomg with the overrides as inline, in a secret and configmap
// THEN helm release for helm module is updated with the updated chart and values in a that namespace
// AND the module status eventually changes to ready
// AND the module status has version as 0.1.1
// AND the helm release values match to that of the updated overrides for new version.
//
// GIVEN an installation of module-operator in a cluster
// AND helm module version 0.1.1 installed in a random namespace with overrides
// WHEN the module is deleted from the that namespace
// THEN helm release for helm module is removed from that namespace
// AND the module is removed from that namespace
func (suite *HelmModuleLifecycleTestSuite) TestOverrideOptionsLifecycle() {
	namespace := common.GetRandomNamespace(6)
	testName := "TestOverrideOptionsLifecycle_namespace_" + namespace
	suite.T().Run(testName, func(t *testing.T) {
		t.Parallel()
		exec(common.NewTestContext(t), namespace, suite)
	})
	suite.T().Cleanup(suite.cleanup)
}

func exec(ctx *common.TestContext, namespace string, suite *HelmModuleLifecycleTestSuite) {
	secretName := "override-secret"
	secretKey := "override-key"
	cmName := "override-cm"
	cmKey := "override-key"
	suite.waitForNamespaceCreated(ctx, namespace)

	module := &api.Module{}
	err := common.UnmarshalTestFile(common.TEST_HELM_MODULE_FILE, module)
	ctx.GomegaWithT.Expect(err).NotTo(gomega.HaveOccurred())
	module.SetNamespace(namespace)

	testNamespacesMutex.Lock()
	testNamespaces[namespace] = []string{module.GetName()}
	testNamespacesMutex.Unlock()

	secret := &corev1.Secret{}
	secret.SetName(secretName)
	secret.SetNamespace(namespace)

	cm := &corev1.ConfigMap{}
	cm.SetName(cmName)
	cm.SetNamespace(namespace)

	module.SetNamespace(namespace)
	module.Spec.TargetNamespace = namespace

	overrides := []*api.ValuesFromSource{
		{
			SecretRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secretName,
				},
				Key: secretKey,
			},
		},
		{
			ConfigMapRef: &corev1.ConfigMapKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cmName,
				},
				Key: cmKey,
			},
		},
	}

	secret, cm, module = mutateAndVerifyModule(suite, ctx, module, common.TEST_HELM_MODULE_OVERRIDE_010_3, secret, common.TEST_HELM_MODULE_OVERRIDE_010_1, cm, common.TEST_HELM_MODULE_OVERRIDE_010_2, overrides, false, false)
	secret, cm, module = mutateAndVerifyModule(suite, ctx, module, common.TEST_HELM_MODULE_OVERRIDE_010_4, secret, common.TEST_HELM_MODULE_OVERRIDE_010_6, cm, common.TEST_HELM_MODULE_OVERRIDE_010_5, overrides, true, false)

	module.Spec.Version = common.TEST_HELM_MODULE_VERSION_011
	secret, cm, module = mutateAndVerifyModule(suite, ctx, module, common.TEST_HELM_MODULE_OVERRIDE_011, secret, common.TEST_HELM_MODULE_OVERRIDE_011_1, cm, common.TEST_HELM_MODULE_OVERRIDE_011_2, overrides, true, true)

	err = ctx.CoreV1Client().Secrets(module.GetNamespace()).Delete(context.TODO(), secret.GetName(), v1.DeleteOptions{})
	ctx.GomegaWithT.Expect(err).NotTo(gomega.HaveOccurred())
	err = ctx.CoreV1Client().ConfigMaps(module.GetNamespace()).Delete(context.TODO(), cm.GetName(), v1.DeleteOptions{})
	ctx.GomegaWithT.Expect(err).NotTo(gomega.HaveOccurred())
	suite.removeModuleAndNamespace(ctx, module)
}

func mutateAndVerifyModule(suite *HelmModuleLifecycleTestSuite, ctx *common.TestContext, module *api.Module, moduleOverrideFile string, secret *corev1.Secret, secretOverrideFile string, cm *corev1.ConfigMap, cmOverrideFile string, overrides []*api.ValuesFromSource, update bool, upgrade bool) (*corev1.Secret, *corev1.ConfigMap, *api.Module) {
	secretData, err := common.LoadTestFile(secretOverrideFile)
	ctx.GomegaWithT.Expect(err).NotTo(gomega.HaveOccurred())
	secret.Data = map[string][]byte{
		overrides[0].SecretRef.Key: secretData,
	}

	cmData, err := common.LoadTestFile(cmOverrideFile)
	ctx.GomegaWithT.Expect(err).NotTo(gomega.HaveOccurred())
	cm.Data = map[string]string{
		overrides[1].ConfigMapRef.Key: string(cmData),
	}

	ctx.GomegaWithT.Expect(err).NotTo(gomega.HaveOccurred())
	if update {
		secret, err = ctx.CoreV1Client().Secrets(module.GetNamespace()).Update(context.TODO(), secret, v1.UpdateOptions{})
	} else {
		secret, err = ctx.CoreV1Client().Secrets(module.GetNamespace()).Create(context.TODO(), secret, v1.CreateOptions{})
	}

	ctx.GomegaWithT.Expect(err).NotTo(gomega.HaveOccurred())

	if update {
		cm, err = ctx.CoreV1Client().ConfigMaps(module.GetNamespace()).Update(context.TODO(), cm, v1.UpdateOptions{})
	} else {
		cm, err = ctx.CoreV1Client().ConfigMaps(module.GetNamespace()).Create(context.TODO(), cm, v1.CreateOptions{})
	}

	ctx.GomegaWithT.Expect(err).NotTo(gomega.HaveOccurred())

	module, overridesFileJSON := suite.createOrUpdateModule(ctx, module, moduleOverrideFile, update, overrides...)

	var secretOverrideMap, cmOverrideMap map[string]interface{}
	err = yaml.Unmarshal(secretData, &secretOverrideMap)
	ctx.GomegaWithT.Expect(err).NotTo(gomega.HaveOccurred())
	err = yaml.Unmarshal(cmData, &cmOverrideMap)
	ctx.GomegaWithT.Expect(err).NotTo(gomega.HaveOccurred())

	if upgrade {
		module = suite.waitForModuleToBeUpgraded(ctx, module)
	}

	return secret, cm, suite.verifyModule(ctx, module, overridesFileJSON, &secretOverrideMap, &cmOverrideMap)
}
