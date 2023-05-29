// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package helm_module_lifecycle_test

import (
	"testing"

	"github.com/verrazzano/verrazzano-modules/tests/common"
)

func (suite *HelmModuleLifecycleTestSuite) TestSingleInstanceLifecycle() {
	testName := "TestSingleInstanceLifecycle_namespace_default"
	suite.T().Run(testName, func(t *testing.T) {
		t.Parallel()
		suite.executeModuleLifecycleOperations(common.DEFAULT_NS)
	})
	suite.T().Cleanup(suite.cleanup)
}
