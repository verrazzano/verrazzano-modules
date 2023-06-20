// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package interrupt

import (
	"fmt"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	api "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/tests/common"
)

func (suite *HelmModuleInterruptTestSuite) TestSingleUpgradeWhileReconciling() {
	ctx := common.NewTestContext(suite.T())

	module := &api.Module{}
	err := common.UnmarshalTestFile(common.TEST_HELM_MODULE_FILE, module)
	ctx.GomegaWithT.Expect(err).ShouldNot(HaveOccurred())

	// GIVEN a Module resource
	// WHEN the Module is created in the cluster
	// THEN the Module Ready condition eventually is true and the installed Helm release has the expected values
	module.Namespace = "default"
	module.Spec.Version = "0.1.0"
	module.Spec.Values = &apiextensionsv1.JSON{Raw: []byte(fmt.Sprintf(`{"deployment": {"delaySeconds": 1},"%s": "%s"}`, customKey, fredValue))}
	suite.createModule(ctx, module)

	version := suite.waitForModuleReadyCondition(ctx, module, corev1.ConditionTrue)
	ctx.GomegaWithT.Expect(version).Should(Equal("0.1.0"))
	suite.verifyHelmValues(ctx, module.Name, module.Namespace, fredValue)

	// GIVEN a Module resource version is being upgraded
	// WHEN the Module is updated while it's upgrading
	// THEN the Module Ready condition eventually is true and the installed Helm release has the expected values

	// Note that we modify the delaySeconds so that the deployment gets updated and pods are rolled out
	module.Spec.Version = "0.1.1"
	module.Spec.Values = &apiextensionsv1.JSON{Raw: []byte(fmt.Sprintf(`{"deployment": {"delaySeconds": 6},"%s": "%s"}`, customKey, fredValue))}
	ctx.T.Logf("Upgrading module to version %s", module.Spec.Version)
	suite.updateModule(ctx, module)

	// Wait for reconciling to begin
	suite.waitForModuleReadyCondition(ctx, module, corev1.ConditionFalse)

	module.Spec.Values = &apiextensionsv1.JSON{Raw: []byte(fmt.Sprintf(`{"deployment": {"delaySeconds": 5},"%s": "%s"}`, customKey, dinoValue))}
	suite.updateModule(ctx, module)

	version = suite.waitForModuleReadyCondition(ctx, module, corev1.ConditionTrue)
	ctx.GomegaWithT.Expect(version).Should(Equal("0.1.1"))
	suite.verifyHelmValues(ctx, module.Name, module.Namespace, dinoValue)

	// GIVEN a Module resource
	// WHEN the Module is deleted
	// THEN the Module is removed from the cluster and the Helm release is uninstalled
	suite.deleteModule(ctx, module)
	suite.verifyModuleAndHelmReleaseDeleted(ctx, module)
}
