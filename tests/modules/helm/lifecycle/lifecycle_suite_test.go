// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package helm_module_lifecycle_test

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/stretchr/testify/suite"
)

type HelmModuleLifecycleTestSuite struct {
	suite.Suite
	gomega types.Gomega
}

func TestHelmModuleLifecycleTestSuite(t *testing.T) {
	helmModuleLifecyclreTestingSuite := new(HelmModuleLifecycleTestSuite)
	helmModuleLifecyclreTestingSuite.gomega = gomega.NewGomegaWithT(t)
	suite.Run(t, helmModuleLifecyclreTestingSuite)
}
