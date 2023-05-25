// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package operator_test

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/stretchr/testify/suite"
)

type OperatorTestSuite struct {
	suite.Suite
	gomega types.Gomega
}

func TestOperatorTestSuite(t *testing.T) {
	operatorTestingSuite := new(OperatorTestSuite)
	operatorTestingSuite.gomega = gomega.NewGomegaWithT(t)
	suite.Run(t, operatorTestingSuite)
}
