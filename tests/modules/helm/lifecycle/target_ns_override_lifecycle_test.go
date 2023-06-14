// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package helm_module_lifecycle_test

import (
	"fmt"
	"testing"

	"github.com/verrazzano/verrazzano-modules/tests/common"
)

// TestTargetNamespaceOverrideLifecycle tests the module lifecycle of a module CR with targetNamespace different from namespace.
// GIVEN an installation of module-operator in a cluster
// WHEN helm module version 0.1.0 is installed in a random namespace with overrides and targetNamespace different than CR
// THEN the helm release for helm module is created in namespace specified by targetNamespace
// AND the module status eventually changes to ready
// AND the helm release values match to that of the overrides.
//
// GIVEN an installation of module-operator in a cluster
// AND helm module version 0.1.0 installed in a random namespace with overrides and targetNamespace different than CR
// WHEN overrides are updated for the installed module in the random namespace
// THEN the values for helm release for helm module is updated in namespace specified by targetNamespace
// AND the module status eventually changes to ready
// AND the helm release values match to that of the updated overrides.
//
// GIVEN an installation of module-operator in a cluster
// AND helm module version 0.1.0 installed in a random namespace with overrides and targetNamespace different than CR
// WHEN helm module version is updated to 0.1.1 alomg with the overrides for the installed module in the random namespace
// THEN helm release for helm module is updated with the updated chart and values in namespace specified by targetNamespace
// AND the module status eventually changes to ready
// AND the module status has version as 0.1.1
// AND the helm release values match to that of the updated overrides for new version.
//
// GIVEN an installation of module-operator in a cluster
// AND helm module version 0.1.1 installed in a random namespace with overrides and targetNamespace different than CR
// WHEN the module is deleted from the random namespace
// THEN helm release for helm module is removed from the namespace specified by targetNamespace
// AND the module is removed from the random namespace
func (suite *HelmModuleLifecycleTestSuite) TestTargetNamespaceOverrideLifecycle() {
	namespace := common.GetRandomNamespace(5)
	targetNamespace := common.GetRandomNamespace(6)
	testName := fmt.Sprintf("TestSingleInstanceLifecycle_namespace_%s_target_ns_%s", namespace, targetNamespace)
	suite.T().Run(testName, func(t *testing.T) {
		t.Parallel()
		suite.executeModuleLifecycleOperationsWithTargetNS(common.NewTestContext(t), namespace, targetNamespace)
	})
	suite.T().Cleanup(suite.cleanup)
}
