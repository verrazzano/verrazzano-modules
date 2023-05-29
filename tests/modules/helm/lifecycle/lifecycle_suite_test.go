// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package helm_module_lifecycle_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/stretchr/testify/suite"
	"github.com/verrazzano/verrazzano-modules/module-operator/clientset/versioned/typed/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/pkg/helm"
	"github.com/verrazzano/verrazzano-modules/pkg/k8sutil"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	"github.com/verrazzano/verrazzano-modules/tests/common"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

type HelmModuleLifecycleTestSuite struct {
	suite.Suite
	gomega types.Gomega
	log    vzlog.VerrazzanoLogger
}

var testNamespaces = make(map[string][]string)

func TestHelmModuleLifecycleTestSuite(t *testing.T) {
	helmModuleLifecyclreTestingSuite := new(HelmModuleLifecycleTestSuite)
	helmModuleLifecyclreTestingSuite.gomega = gomega.NewGomegaWithT(t)
	helmModuleLifecyclreTestingSuite.log = vzlog.DefaultLogger()
	suite.Run(t, helmModuleLifecyclreTestingSuite)
}

func (suite *HelmModuleLifecycleTestSuite) executeModuleLifecycleOperations(namespace string) {
	err := common.WaitForNamespaceCreated(namespace)
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	testNamespaces[namespace] = []string{""}
	module := &api.Module{}
	err = common.UnmarshalTestFile(common.TEST_HELM_MODULE_FILE, module)
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	testNamespaces[namespace] = append(testNamespaces[namespace], module.GetName())
	c := suite.getModuleClient()

	module.SetNamespace(namespace)
	module.Spec.TargetNamespace = namespace
	module, overrides := suite.createOrUpdateModule(c, module, common.TEST_HELM_MODULE_OVERRIDE_010, false)
	module = suite.verifyModule(c, module, overrides)

	module, overrides = suite.createOrUpdateModule(c, module, common.TEST_HELM_MODULE_OVERRIDE_010_1, true)
	module = suite.verifyModule(c, module, overrides)

	module.Spec.Version = common.TEST_HELM_MODULE_VERSION_011
	module, overrides = suite.createOrUpdateModule(c, module, common.TEST_HELM_MODULE_OVERRIDE_011, true)
	module, err = common.WaitForModuleToBeUpgraded(c, module)
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	module = suite.verifyModule(c, module, overrides)

	suite.log.Infof("\ndelete module %s, version %s, namespace %s\n", module.GetName(), module.Spec.Version, module.GetNamespace())
	c.Modules(module.GetNamespace()).Delete(context.TODO(), module.GetName(), v1.DeleteOptions{})
	suite.verifyModuleDeleted(c, module)
	corev1client, err := k8sutil.GetCoreV1Client()
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	corev1client.Namespaces().Delete(context.TODO(), namespace, v1.DeleteOptions{})
}

func (suite *HelmModuleLifecycleTestSuite) cleanup() {
	corev1client, err := k8sutil.GetCoreV1Client()
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	c := suite.getModuleClient()
	for namespace, modules := range testNamespaces {
		if suite.deleteNamespace(namespace) {
			corev1client.Namespaces().Delete(context.TODO(), namespace, v1.DeleteOptions{})
			common.WaitForNamespaceDeleted(namespace)
		} else {
			for _, moduleName := range modules {
				c.Modules(namespace).Delete(context.TODO(), moduleName, v1.DeleteOptions{})
				common.WaitForModuleToBeDeleted(c, namespace, moduleName)
			}
		}

	}
}

func (suite *HelmModuleLifecycleTestSuite) createOrUpdateModule(c *v1alpha1.PlatformV1alpha1Client, module *api.Module, overridesFile string, update bool) (*api.Module, *apiextensionsv1.JSON) {
	var err error
	op := "create"
	if update {
		op = "update"
	}

	suite.log.Infof("\n%s module %s, version %s, namespace %s", op, module.GetName(), module.Spec.Version, module.GetNamespace())
	overrides := suite.generateOverridesFromFile(overridesFile)
	module.Spec.Overrides = []api.Overrides{
		{
			Values: overrides,
		},
	}

	if update {
		module, err = c.Modules(module.GetNamespace()).Update(context.TODO(), module, v1.UpdateOptions{})
	} else {
		module, err = c.Modules(module.GetNamespace()).Create(context.TODO(), module, v1.CreateOptions{})
	}

	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return module, overrides
}

func (suite *HelmModuleLifecycleTestSuite) verifyModule(c *v1alpha1.PlatformV1alpha1Client, module *api.Module, overrides *apiextensionsv1.JSON) *api.Module {
	deployedModule := suite.verifyModuleIsReady(c, module)
	suite.verifyHelmReleaseStatus(c, module, deployedModule)
	suite.verifyHelmValues(c, module, deployedModule, overrides)
	deployedModule, err := c.Modules(module.GetNamespace()).Get(context.TODO(), module.GetName(), v1.GetOptions{})
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return deployedModule
}

func (suite *HelmModuleLifecycleTestSuite) verifyModuleIsReady(c *v1alpha1.PlatformV1alpha1Client, module *api.Module) *api.Module {
	deployedModule, err := common.WaitForModuleToBeReady(c, module)
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	suite.gomega.Expect(deployedModule.Status.Version).To(gomega.Equal(module.Spec.Version))
	return deployedModule
}

func (suite *HelmModuleLifecycleTestSuite) verifyHelmReleaseStatus(c *v1alpha1.PlatformV1alpha1Client, module *api.Module, deployedModule *api.Module) {
	status, err := helm.GetHelmReleaseStatus(deployedModule.GetName(), deployedModule.GetNamespace())
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	suite.gomega.Expect(status).To(gomega.Equal(helm.ReleaseStatusDeployed))
	suite.gomega.Expect(deployedModule.Status.Version).To(gomega.Equal(module.Spec.Version))
}

func (suite *HelmModuleLifecycleTestSuite) verifyHelmValues(c *v1alpha1.PlatformV1alpha1Client, module *api.Module, deployedModule *api.Module, overrides *apiextensionsv1.JSON) {
	_ = common.Retry(common.DefaultRetry, vzlog.DefaultLogger(), true, func() (bool, error) {
		deployedValues, err := helm.GetValuesMap(vzlog.DefaultLogger(), deployedModule.GetName(), deployedModule.GetNamespace())
		suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
		deployedValuesBytes, err := json.Marshal(deployedValues)
		suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
		appliedValuesBytes, err := overrides.MarshalJSON()
		suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
		return bytes.Equal(deployedValuesBytes, appliedValuesBytes), nil
	})
}

func (suite *HelmModuleLifecycleTestSuite) generateOverridesFromFile(overridesFile string) *apiextensionsv1.JSON {
	overrides := &apiextensionsv1.JSON{}
	err := common.UnmarshalTestFile(overridesFile, overrides)
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return overrides
}

func (suite *HelmModuleLifecycleTestSuite) getModuleClient() *v1alpha1.PlatformV1alpha1Client {
	config, err := k8sutil.GetKubeConfig()
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	c, err := v1alpha1.NewForConfig(config)
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return c
}

func (suite *HelmModuleLifecycleTestSuite) verifyModuleDeleted(c *v1alpha1.PlatformV1alpha1Client, module *api.Module) {
	err := common.WaitForModuleToBeDeleted(c, module.GetNamespace(), module.GetName())
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	helmReleaseInstalled, err := helm.IsReleaseInstalled(module.GetName(), module.GetNamespace())
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	suite.gomega.Expect(helmReleaseInstalled).To(gomega.BeFalse())
}

func (suite *HelmModuleLifecycleTestSuite) deleteNamespace(ns string) bool {
	return strings.HasPrefix(ns, "ns")
}
