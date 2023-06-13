// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package operator_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// OperatorTestSuite is the test suite for module-operator tests.
type OperatorTestSuite struct {
	suite.Suite
}

// TestOperatorTestSuite is the driver test for module-operator tests.
func TestOperatorTestSuite(t *testing.T) {
	operatorTestingSuite := new(OperatorTestSuite)
	suite.Run(t, operatorTestingSuite)
}
