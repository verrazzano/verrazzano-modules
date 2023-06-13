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
	"github.com/verrazzano/verrazzano-modules/module-operator/clientset/versioned/typed/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/status"
	"github.com/verrazzano/verrazzano-modules/pkg/helm"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	"github.com/verrazzano/verrazzano-modules/tests/common"
)

const (
	shortWaitTimeout     = 3 * time.Minute
	shortPollingInterval = 2 * time.Second

	customKey = "customKey"
	fredValue = "fred"
	dinoValue = "dino"
)

// TestUpdateWhileReconciling tests updating a Module while it's reconciling and validates that all updates
// are applied correctly.
func (suite *HelmModuleInterruptTestSuite) TestUpdateWhileReconciling() {
	c, err := common.GetModuleClient()
	suite.gomega.Expect(err).ShouldNot(HaveOccurred())

	module := &api.Module{}
	err = common.UnmarshalTestFile(common.TEST_HELM_MODULE_FILE, module)
	suite.gomega.Expect(err).ShouldNot(HaveOccurred())

	// GIVEN a Module resource is created and the Ready condition is false
	// WHEN the Module values are updated
	// THEN the Module Ready condition eventually is true and the installed Helm release has the expected values
	module.Namespace = "default"
	module.Spec.Values = &apiextensionsv1.JSON{Raw: []byte(`{"deployment": {"delaySeconds": 10}}`)}
	fmt.Printf("Installing module %s/%s\n", module.Namespace, module.Name)
	suite.createModule(c, module)

	fmt.Println("Waiting for ready condition false")
	suite.waitForModuleReadyCondition(c, module, corev1.ConditionFalse)

	// Note that we modify the delaySeconds so that the deployment gets updated and pods are rolled out
	module.Spec.Values = &apiextensionsv1.JSON{Raw: []byte(fmt.Sprintf(`{"deployment": {"delaySeconds": 11},"%s": "%s"}`, customKey, fredValue))}
	fmt.Println("Updating module values")
	suite.updateModule(c, module)

	fmt.Println("Waiting for ready condition true and verifying helm values")
	version := suite.waitForModuleReadyCondition(c, module, corev1.ConditionTrue)
	suite.verifyHelmValues(module.Name, module.Namespace, fredValue)

	// GIVEN a Module resource is updated and the Ready condition is false
	// WHEN the Module values are updated again
	// THEN the Module Ready condition eventually is true and the installed Helm release has the expected values
	module.Spec.Values = &apiextensionsv1.JSON{Raw: []byte(fmt.Sprintf(`{"deployment": {"delaySeconds": 12},"%s": "barney"}`, customKey))}
	fmt.Println("Updating module values")
	suite.updateModule(c, module)

	fmt.Println("Waiting for ready condition false")
	suite.waitForModuleReadyCondition(c, module, corev1.ConditionFalse)

	module.Spec.Values = &apiextensionsv1.JSON{Raw: []byte(fmt.Sprintf(`{"deployment": {"delaySeconds": 13},"%s": "%s"}`, customKey, dinoValue))}
	fmt.Println("Updating module values")
	suite.updateModule(c, module)

	fmt.Println("Waiting for ready condition true and verifying helm values")
	newVersion := suite.waitForModuleReadyCondition(c, module, corev1.ConditionTrue)
	suite.gomega.Expect(newVersion).Should(Equal(version), "Expected the module version to not have changed")
	suite.verifyHelmValues(module.Name, module.Namespace, dinoValue)

	// GIVEN a Module resource
	// WHEN the Module is deleted
	// THEN the Module is removed from the cluster and the Helm release is uninstalled
	fmt.Println("Deleting module")
	suite.deleteModule(c, module)
	suite.verifyModuleAndHelmReleaseDeleted(c, module)
}

func (suite *HelmModuleInterruptTestSuite) createModule(c *v1alpha1.PlatformV1alpha1Client, module *api.Module) {
	suite.gomega.Eventually(func() error {
		_, err := c.Modules(module.GetNamespace()).Create(context.TODO(), module, v1.CreateOptions{})
		return err
	}, shortWaitTimeout, shortPollingInterval).ShouldNot(HaveOccurred())
}

func (suite *HelmModuleInterruptTestSuite) updateModule(c *v1alpha1.PlatformV1alpha1Client, module *api.Module) {
	name := module.Name
	namespace := module.Namespace
	version := module.Spec.Version
	values := module.Spec.Values

	suite.gomega.Eventually(func() error {
		var err error

		// Get the latest Module or else the code will never resolve conflicts
		module, err = c.Modules(namespace).Get(context.TODO(), name, v1.GetOptions{})
		if err != nil {
			return err
		}

		// Update the fetched module with values from the passed in module
		if version != "" {
			module.Spec.Version = version
		}
		module.Spec.Values = values

		_, err = c.Modules(module.GetNamespace()).Update(context.TODO(), module, v1.UpdateOptions{})
		return err
	}, shortWaitTimeout, shortPollingInterval).ShouldNot(HaveOccurred())
}

func (suite *HelmModuleInterruptTestSuite) waitForModuleReadyCondition(c *v1alpha1.PlatformV1alpha1Client, module *api.Module, expectedStatus corev1.ConditionStatus) string {
	var version string

	suite.gomega.Eventually(func() (corev1.ConditionStatus, error) {
		module, err := c.Modules(module.GetNamespace()).Get(context.TODO(), module.GetName(), v1.GetOptions{})
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

func (suite *HelmModuleInterruptTestSuite) verifyHelmValues(releaseName, releaseNamespace, value string) {
	deployedValues, err := helm.GetValuesMap(vzlog.DefaultLogger(), releaseName, releaseNamespace)
	suite.gomega.Expect(err).ShouldNot(HaveOccurred())
	suite.gomega.Expect(deployedValues[customKey]).To(BeEquivalentTo(value))
}

func (suite *HelmModuleInterruptTestSuite) deleteModule(c *v1alpha1.PlatformV1alpha1Client, module *api.Module) {
	suite.gomega.Eventually(func() error {
		return c.Modules(module.Namespace).Delete(context.TODO(), module.Name, v1.DeleteOptions{})
	}, shortWaitTimeout, shortPollingInterval).ShouldNot(HaveOccurred())
}

func (suite *HelmModuleInterruptTestSuite) verifyModuleAndHelmReleaseDeleted(c *v1alpha1.PlatformV1alpha1Client, module *api.Module) {
	suite.waitForModuleToBeDeleted(c, module.GetNamespace(), module.GetName())
	helmReleaseInstalled, err := helm.IsReleaseInstalled(module.GetName(), module.GetNamespace())
	suite.gomega.Expect(err).ShouldNot(HaveOccurred())
	suite.gomega.Expect(helmReleaseInstalled).To(BeFalse())
}

func (suite *HelmModuleInterruptTestSuite) waitForModuleToBeDeleted(c *v1alpha1.PlatformV1alpha1Client, namespace string, name string) {
	suite.gomega.Eventually(func() error {
		_, err := c.Modules(namespace).Get(context.TODO(), name, v1.GetOptions{})
		return err
	}, shortWaitTimeout, shortPollingInterval).Should(MatchError(MatchRegexp("not found")))
}
