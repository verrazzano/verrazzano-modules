// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package helm_module_lifecycle_test

import (
	"fmt"
	"testing"
)

// TestMultipleInstanceLifecycle tests the module lifecycle of multiple module CRs concurrently.
// GIVEN an installation of module-operator in a cluster
// WHEN helm module version 0.1.0 is installed in multiple namespaces with overrides
// THEN the helm release for helm module is created in corresponding namespaces
// AND the module status eventually changes to ready
// AND the helm release values match to that of the overrides.
//
// GIVEN an installation of module-operator in a cluster
// AND helm module version 0.1.0 installed in multiple namespaces with overrides
// WHEN overrides are updated for the installed modules in corresponding namespaces
// THEN the values for helm release for helm module is updated in corresponding namespaces
// AND the module status eventually changes to ready
// AND the helm release values match to that of the updated overrides.
//
// GIVEN an installation of module-operator in a cluster
// AND helm module version 0.1.0 installed in multiple namespaces with overrides
// WHEN helm module version is updated to 0.1.1 alomg with the overrides for the installed modules in corresponding namespaces
// THEN helm release for helm module is updated with the updated chart and values in corresponding namespaces
// AND the module status eventually changes to ready
// AND the module status has version as 0.1.1
// AND the helm release values match to that of the updated overrides for new version.
//
// GIVEN an installation of module-operator in a cluster
// AND helm module version 0.1.1 installed in multiple namespaces with overrides
// WHEN the module is deleted from the corresponding namespaces
// THEN helm release for helm module is removed from the corresponding namespaces
// AND the module is removed from the corresponding namespaces
func (suite *HelmModuleLifecycleTestSuite) TestMultipleInstanceLifecycle() {
	for count := 0; count < 20; count++ {
		namespace := fmt.Sprintf("ns%v", count)
		testName := fmt.Sprintf("TestMultipleInstanceLifecycle_namespace_%s", namespace)
		suite.T().Run(testName, func(t *testing.T) {
			t.Parallel()
			suite.executeModuleLifecycleOperations(testName, namespace)
		})
		suite.T().Cleanup(suite.cleanup)
	}
}
