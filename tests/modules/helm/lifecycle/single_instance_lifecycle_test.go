// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package helm_module_lifecycle_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/onsi/gomega"
	api "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/clientset/versioned/typed/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/pkg/helm"
	"github.com/verrazzano/verrazzano-modules/pkg/k8sutil"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	"github.com/verrazzano/verrazzano-modules/tests/common"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (suite *HelmModuleLifecycleTestSuite) TestSingleInstanceLifecycle() {
	module := &api.Module{}
	err := common.UnmarshalTestFile(common.TEST_HELM_MODULE_FILE, module)
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	overrides := &apiextensionsv1.JSON{}
	err = common.UnmarshalTestFile(common.TEST_HELM_MODULE_OVERRIDE_010, overrides)
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	module.Spec.Overrides = []api.Overrides{
		{
			Values: overrides,
		},
	}
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	config, err := k8sutil.GetKubeConfig()
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	c, err := v1alpha1.NewForConfig(config)
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	_, err = c.Modules(common.DEFAULT_NS).Create(context.TODO(), module, v1.CreateOptions{})
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())

	deployedModule, err := common.WaitForModuleToBeReady(c, common.DEFAULT_NS, module.GetName())
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	suite.gomega.Expect(deployedModule.Status.Version).To(gomega.Equal(module.Spec.Version))

	status, err := helm.GetHelmReleaseStatus(deployedModule.GetName(), common.DEFAULT_NS)
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	suite.gomega.Expect(status).To(gomega.Equal(helm.ReleaseStatusDeployed))

	deployedValues, err := helm.GetValuesMap(vzlog.DefaultLogger(), deployedModule.GetName(), common.DEFAULT_NS)
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	deployedValuesBytes, err := json.Marshal(deployedValues)
	fmt.Println(string(deployedValuesBytes))
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	appliedValuesBytes, err := overrides.MarshalJSON()
	fmt.Println(string(appliedValuesBytes))
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	suite.gomega.Expect(deployedValuesBytes).To(gomega.Equal(appliedValuesBytes))
}
