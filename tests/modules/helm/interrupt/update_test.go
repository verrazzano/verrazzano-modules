// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package interrupt

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/status"
	"github.com/verrazzano/verrazzano-modules/pkg/helm"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	"github.com/verrazzano/verrazzano-modules/tests/common"
)

const (
	shortWaitTimeout     = 3 * time.Minute
	shortPollingInterval = 2 * time.Second

	customKey   = "customKey"
	fredValue   = "fred"
	barneyValue = "barney"
	dinoValue   = "dino"
)

// TestUpdateWhileReconciling tests updating a Module while it's reconciling and validates that all updates
// are applied correctly.
func (suite *HelmModuleInterruptTestSuite) TestUpdateWhileReconciling() {
	ctx := common.NewTestContext(suite.T())

	module := &api.Module{}
	err := common.UnmarshalTestFile(common.TEST_HELM_MODULE_FILE, module)
	ctx.GomegaWithT.Expect(err).ShouldNot(HaveOccurred())

	// GIVEN a Module resource is created and the Ready condition is false
	// WHEN the Module values are updated
	// THEN the Module Ready condition eventually is true and the installed Helm release has the expected values
	module.Namespace = "default"
	module.Spec.Values = &apiextensionsv1.JSON{Raw: []byte(`{"deployment": {"delaySeconds": 6}}`)}
	suite.createModule(ctx, module)

	suite.waitForModuleReadyCondition(ctx, module, corev1.ConditionFalse)

	// Note that we modify the delaySeconds so that the deployment gets updated and pods are rolled out
	module.Spec.Values = &apiextensionsv1.JSON{Raw: []byte(fmt.Sprintf(`{"deployment": {"delaySeconds": 7},"%s": "%s"}`, customKey, fredValue))}
	suite.updateModule(ctx, module)

	version := suite.waitForModuleReadyCondition(ctx, module, corev1.ConditionTrue)
	suite.verifyHelmValues(ctx, module.Name, module.Namespace, fredValue)

	// GIVEN a Module resource is updated and the Ready condition is false
	// WHEN the Module values are updated again
	// THEN the Module Ready condition eventually is true and the installed Helm release has the expected values
	module.Spec.Values = &apiextensionsv1.JSON{Raw: []byte(fmt.Sprintf(`{"deployment": {"delaySeconds": 8},"%s": "%s"}`, customKey, barneyValue))}
	suite.updateModule(ctx, module)

	suite.waitForModuleReadyCondition(ctx, module, corev1.ConditionFalse)

	module.Spec.Values = &apiextensionsv1.JSON{Raw: []byte(fmt.Sprintf(`{"deployment": {"delaySeconds": 6},"%s": "%s"}`, customKey, dinoValue))}
	suite.updateModule(ctx, module)

	newVersion := suite.waitForModuleReadyCondition(ctx, module, corev1.ConditionTrue)
	ctx.GomegaWithT.Expect(newVersion).Should(Equal(version), "Expected the module version to not have changed")
	suite.verifyHelmValues(ctx, module.Name, module.Namespace, dinoValue)

	// GIVEN a Module resource
	// WHEN the Module is deleted
	// THEN the Module is removed from the cluster and the Helm release is uninstalled
	suite.deleteModule(ctx, module)
	suite.verifyModuleAndHelmReleaseDeleted(ctx, module)
}

func (suite *HelmModuleInterruptTestSuite) createModule(ctx *common.TestContext, module *api.Module) {
	ctx.T.Logf("Installing module %s/%s", module.Namespace, module.Name)
	ctx.GomegaWithT.Eventually(func() error {
		_, err := ctx.ModuleClient().Modules(module.GetNamespace()).Create(context.TODO(), module, v1.CreateOptions{})
		return err
	}, shortWaitTimeout, shortPollingInterval).ShouldNot(HaveOccurred())
}

func (suite *HelmModuleInterruptTestSuite) updateModule(ctx *common.TestContext, module *api.Module) {
	ctx.T.Log("Updating module values")
	name := module.Name
	namespace := module.Namespace
	version := module.Spec.Version
	values := module.Spec.Values

	ctx.GomegaWithT.Eventually(func() error {
		var err error

		// Get the latest Module or else the code will never resolve conflicts
		module, err = ctx.ModuleClient().Modules(namespace).Get(context.TODO(), name, v1.GetOptions{})
		if err != nil {
			return err
		}

		// Update the fetched module with values from the passed in module
		if version != "" {
			module.Spec.Version = version
		}
		module.Spec.Values = values

		_, err = ctx.ModuleClient().Modules(module.GetNamespace()).Update(context.TODO(), module, v1.UpdateOptions{})
		return err
	}, shortWaitTimeout, shortPollingInterval).ShouldNot(HaveOccurred())
}

func (suite *HelmModuleInterruptTestSuite) waitForModuleReadyCondition(ctx *common.TestContext, module *api.Module, expectedStatus corev1.ConditionStatus) string {
	var version string

	ctx.T.Logf("Waiting for module ready condition to be: %v", expectedStatus)
	ctx.GomegaWithT.Eventually(func() (corev1.ConditionStatus, error) {
		module, err := ctx.ModuleClient().Modules(module.GetNamespace()).Get(context.TODO(), module.GetName(), v1.GetOptions{})
		if err != nil {
			return "", err
		}

		version = module.Status.LastSuccessfulVersion
		cond := status.GetReadyCondition(module)
		if cond == nil {
			return "", nil
		}
		return cond.Status, nil
	}, shortWaitTimeout, shortPollingInterval).Should(Equal(expectedStatus))

	return version
}

func (suite *HelmModuleInterruptTestSuite) verifyHelmValues(ctx *common.TestContext, releaseName, releaseNamespace, value string) {
	ctx.T.Log("Verifying helm release values")
	deployedValues, err := helm.GetValuesMap(vzlog.DefaultLogger(), releaseName, releaseNamespace)
	ctx.GomegaWithT.Expect(err).ShouldNot(HaveOccurred())
	ctx.GomegaWithT.Expect(deployedValues[customKey]).To(BeEquivalentTo(value))
}

func (suite *HelmModuleInterruptTestSuite) deleteModule(ctx *common.TestContext, module *api.Module) {
	ctx.T.Logf("Deleting module %s/%s", module.Namespace, module.Name)
	ctx.GomegaWithT.Eventually(func() error {
		return ctx.ModuleClient().Modules(module.Namespace).Delete(context.TODO(), module.Name, v1.DeleteOptions{})
	}, shortWaitTimeout, shortPollingInterval).ShouldNot(HaveOccurred())
}

func (suite *HelmModuleInterruptTestSuite) verifyModuleAndHelmReleaseDeleted(ctx *common.TestContext, module *api.Module) {
	suite.waitForModuleToBeDeleted(ctx, module.GetNamespace(), module.GetName())
	helmReleaseInstalled, err := helm.IsReleaseInstalled(module.GetName(), module.GetNamespace())
	ctx.GomegaWithT.Expect(err).ShouldNot(HaveOccurred())
	ctx.GomegaWithT.Expect(helmReleaseInstalled).To(BeFalse())
}

func (suite *HelmModuleInterruptTestSuite) waitForModuleToBeDeleted(ctx *common.TestContext, namespace string, name string) {
	ctx.T.Log("Waiting for module to be deleted")
	ctx.GomegaWithT.Eventually(func() error {
		_, err := ctx.ModuleClient().Modules(namespace).Get(context.TODO(), name, v1.GetOptions{})
		return err
	}, shortWaitTimeout, shortPollingInterval).Should(MatchError(MatchRegexp("not found")))
}
