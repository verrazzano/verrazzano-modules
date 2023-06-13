// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package interrupt

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/stretchr/testify/suite"
)

type HelmModuleInterruptTestSuite struct {
	suite.Suite
	gomega types.Gomega
}

// TestHelmModuleInterruptTestSuite runs the interrupt tests for the helm module.
func TestHelmModuleInterruptTestSuite(t *testing.T) {
	helmModuleInterruptTestSuite := &HelmModuleInterruptTestSuite{gomega: gomega.NewWithT(t)}
	suite.Run(t, helmModuleInterruptTestSuite)
}
