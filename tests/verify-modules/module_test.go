// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package modules_test

import (
	"context"

	"github.com/onsi/gomega"
	api "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/clientset/versioned/typed/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/pkg/k8sutil"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (suite *ModulesTestSuite) TestInstallModule() {
	config, err := k8sutil.GetKubeConfig()
	suite.gomega.Expect(err).Should(gomega.BeNil())
	c, err := v1alpha1.NewForConfig(config)
	suite.gomega.Expect(err).Should(gomega.BeNil())
	module := &api.Module{}
	module.SetName("vz-test")
	module.Spec.ModuleName = "vz-test"
	_, err = c.Modules("default").Create(context.TODO(), module, v1.CreateOptions{})
	suite.gomega.Expect(err).Should(gomega.BeNil())
}
