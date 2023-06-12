// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package helm_module_lifecycle_test

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/verrazzano/verrazzano-modules/tests/common"
)

// TestSingleInstanceLifecycle tests the module lifecycle of a module CR.
// GIVEN an installation of module-operator in a cluster
// WHEN helm module version 0.1.0 is installed in default namespaces with overrides
// THEN the helm release for helm module is created in default namespace
// AND the module status eventually changes to ready
// AND the helm release values match to that of the overrides.
//
// GIVEN an installation of module-operator in a cluster
// AND helm module version 0.1.0 installed in default namespace with overrides
// WHEN overrides are updated for the installed module in default namespace
// THEN the values for helm release for helm module is updated in default namespace
// AND the module status eventually changes to ready
// AND the helm release values match to that of the updated overrides.
//
// GIVEN an installation of module-operator in a cluster
// AND helm module version 0.1.0 installed in default namespace with overrides
// WHEN helm module version is updated to 0.1.1 alomg with the overrides for the installed module in default namespace
// THEN helm release for helm module is updated with the updated chart and values in default namespace
// AND the module status eventually changes to ready
// AND the module status has version as 0.1.1
// AND the helm release values match to that of the updated overrides for new version.
//
// GIVEN an installation of module-operator in a cluster
// AND helm module version 0.1.1 installed in default namespace with overrides
// WHEN the module is deleted from the default namespace
// THEN helm release for helm module is removed from the default namespace
// AND the module is removed from the default namespace
func (suite *HelmModuleLifecycleTestSuite) TestSingleInstanceLifecycle() {
	testName := "TestSingleInstanceLifecycle_namespace_default"
	suite.T().Run(testName, func(t *testing.T) {
		t.Parallel()
		gomegaWithT := gomega.NewWithT(t)
		suite.executeModuleLifecycleOperations(t, gomegaWithT, common.DEFAULT_NS)
	})
	suite.T().Cleanup(suite.cleanup)
}
