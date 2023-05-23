// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package modules_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/stretchr/testify/suite"
	"github.com/verrazzano/verrazzano-modules/pkg/k8sutil"
)

type ModulesTestSuite struct {
	suite.Suite
	gomega types.Gomega
}

func (suite *ModulesTestSuite) SetupSuite() {
	err := k8sutil.NewYAMLApplier(nil, "").ApplyFTDefaultConfig(fmt.Sprintf("%s/%s", os.Getenv("VMO_ROOT"), "build/deploy/verrazzano-module-operator.yaml"), nil)
	suite.gomega.Expect(err).Should(gomega.BeNil())
}

// this function executes after all tests executed
/*func (suite *ModulesTestSuite) TearDownSuite() {
	err := k8sutil.NewYAMLApplier(nil, "").DeleteFTDefaultConfig(fmt.Sprintf("%s/%s", os.Getenv("VMO_ROOT"), "build/deploy/verrazzano-module-operator.yaml"), nil)
	suite.gomega.Expect(err).Should(gomega.BeNil())
}*/

func TestModulesTestSuite(t *testing.T) {
	modulesTestingSuite := new(ModulesTestSuite)
	modulesTestingSuite.gomega = gomega.NewGomegaWithT(t)
	suite.Run(t, modulesTestingSuite)
}
