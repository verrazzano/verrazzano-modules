// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package interrupt

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type HelmModuleInterruptTestSuite struct {
	suite.Suite
	t *testing.T
}

// TestHelmModuleInterruptTestSuite runs the interrupt tests for the helm module.
func TestHelmModuleInterruptTestSuite(t *testing.T) {
	helmModuleInterruptTestSuite := &HelmModuleInterruptTestSuite{t: t}
	suite.Run(t, helmModuleInterruptTestSuite)
}
