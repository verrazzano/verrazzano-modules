// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package helm_module_lifecycle_test

import (
	"fmt"
	"testing"
)

func (suite *HelmModuleLifecycleTestSuite) TestMultipleInstanceLifecycle() {
	for count := 0; count < 20; count++ {
		namespace := fmt.Sprintf("ns%v", count)
		testName := fmt.Sprintf("TestMultipleInstanceLifecycle_namespace_%s", namespace)
		suite.T().Run(testName, func(t *testing.T) {
			t.Parallel()
			suite.executeModuleLifecycleOperations(namespace)
		})
		suite.T().Cleanup(suite.cleanup)
	}
}
