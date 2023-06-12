// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package helm_module_lifecycle_test

import (
	"context"
	"encoding/json"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/stretchr/testify/suite"
	"github.com/verrazzano/verrazzano-modules/module-operator/clientset/versioned/typed/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/status"
	"github.com/verrazzano/verrazzano-modules/pkg/helm"
	"github.com/verrazzano/verrazzano-modules/pkg/k8sutil"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	"github.com/verrazzano/verrazzano-modules/pkg/yaml"
	"github.com/verrazzano/verrazzano-modules/tests/common"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	kerrs "k8s.io/apimachinery/pkg/api/errors"
)

// HelmModuleLifecycleTestSuite is the test suite for the lifecycle tests for the helm module.
type HelmModuleLifecycleTestSuite struct {
	suite.Suite
	gomega types.Gomega
}

var testNamespaces = make(map[string][]string)
var testNamespacesMutex sync.Mutex

const (
	shortWaitTimeout     = 5 * time.Minute
	shortPollingInterval = 2 * time.Second
)

// TestHelmModuleLifecycleTestSuite is the driver test for the lifecycle tests for the helm module.
func TestHelmModuleLifecycleTestSuite(t *testing.T) {
	helmModuleLifecyclreTestingSuite := new(HelmModuleLifecycleTestSuite)
	helmModuleLifecyclreTestingSuite.gomega = gomega.NewGomegaWithT(t)
	suite.Run(t, helmModuleLifecyclreTestingSuite)
}

func (suite *HelmModuleLifecycleTestSuite) executeModuleLifecycleOperations(t *testing.T, namespace string) {
	err := suite.waitForNamespaceCreated(t, namespace)
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	module := &api.Module{}
	err = common.UnmarshalTestFile(common.TEST_HELM_MODULE_FILE, module)
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())

	testNamespacesMutex.Lock()
	testNamespaces[namespace] = []string{module.GetName()}
	testNamespacesMutex.Unlock()

	c := suite.getModuleClient()

	module.SetNamespace(namespace)
	module.Spec.TargetNamespace = namespace
	module, overrides := suite.createOrUpdateModule(t, c, module, common.TEST_HELM_MODULE_OVERRIDE_010, false)
	module = suite.verifyModule(t, c, module, overrides)

	module, overrides = suite.createOrUpdateModule(t, c, module, common.TEST_HELM_MODULE_OVERRIDE_010_1, true)
	module = suite.verifyModule(t, c, module, overrides)

	module.Spec.Version = common.TEST_HELM_MODULE_VERSION_011
	module, overrides = suite.createOrUpdateModule(logger, c, module, common.TEST_HELM_MODULE_OVERRIDE_011, true)
	module = suite.waitForModuleToBeUpgraded(logger, c, module)
	module = suite.verifyModule(logger, c, module, overrides)
	suite.removeModuleAndNamespace(logger, c, module)
}

func (suite *HelmModuleLifecycleTestSuite) cleanup() {
	corev1client, err := k8sutil.GetCoreV1Client()
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	c := suite.getModuleClient()
	for namespace, modules := range testNamespaces {
		if suite.deleteNamespace(namespace) {
			err = corev1client.Namespaces().Delete(context.TODO(), namespace, v1.DeleteOptions{})
			if kerrs.IsNotFound(err) {
				continue
			}

			suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
			err = suite.waitForNamespaceDeleted(suite.T(), namespace)
			suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
		} else {
			for _, moduleName := range modules {
				err = c.Modules(namespace).Delete(context.TODO(), moduleName, v1.DeleteOptions{})
				if kerrs.IsNotFound(err) {
					continue
				}

				suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
				suite.waitForModuleToBeDeleted(suite.T(), c, namespace, moduleName)
			}
		}

	}
}

func (suite *HelmModuleLifecycleTestSuite) createOrUpdateModule(logger testLogger, c *v1alpha1.PlatformV1alpha1Client, module *api.Module, overridesFile string, update bool, otherOverrides ...*api.ValuesFromSource) (*api.Module, *apiextensionsv1.JSON) {
	op := "create"
	if update {
		op = "update"
	}
	name := module.Name
	namespace := module.Namespace
	version := module.Spec.Version

	logger.log("%s module %s, version %s, namespace %s", op, module.GetName(), module.Spec.Version, module.GetNamespace())

	// Build the values and valuesFrom
	values := suite.generateOverridesFromFile(overridesFile)
	var valuesFrom []api.ValuesFromSource
	for _, toAppend := range otherOverrides {
		valuesFrom = append(valuesFrom, *toAppend)
	}
	suite.gomega.Eventually(func() error {
		var err error

		// Get the latest Module or else the code will never resolve conflicts
		if op != "create" {
			module, err = c.Modules(namespace).Get(context.TODO(), name, v1.GetOptions{})
			if err != nil {
				return err
			}
		}
		// Update the version
		if version != "" {
			module.Spec.Version = version
		}

		// Set the module values and valuesFrom
		module.Spec.Values = values
		module.Spec.ValuesFrom = valuesFrom

		// Do the create or update
		if update {
			module, err = c.Modules(module.GetNamespace()).Update(context.TODO(), module, v1.UpdateOptions{})
		} else {
			module, err = c.Modules(module.GetNamespace()).Create(context.TODO(), module, v1.CreateOptions{})
		}

		return err
	}, shortWaitTimeout, shortPollingInterval).ShouldNot(gomega.HaveOccurred())

	return module, values
}

func (suite *HelmModuleLifecycleTestSuite) verifyModule(logger testLogger, c *v1alpha1.PlatformV1alpha1Client, module *api.Module, overrides *apiextensionsv1.JSON, otherOverrides ...*map[string]interface{}) *api.Module {
	deployedModule := suite.verifyModuleIsReady(logger, c, module)
	suite.verifyHelmReleaseStatus(c, module, deployedModule)
	suite.verifyHelmValues(logger, c, module, deployedModule, overrides, otherOverrides...)
	deployedModule, err := c.Modules(module.GetNamespace()).Get(context.TODO(), module.GetName(), v1.GetOptions{})
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return deployedModule
}

func (suite *HelmModuleLifecycleTestSuite) verifyModuleIsReady(logger testLogger, c *v1alpha1.PlatformV1alpha1Client, module *api.Module) *api.Module {
	deployedModule := suite.waitForModuleToBeReady(logger, c, module)
	suite.gomega.Expect(deployedModule.Status.LastSuccessfulVersion).To(gomega.Equal(module.Spec.Version))
	return deployedModule
}

func (suite *HelmModuleLifecycleTestSuite) verifyHelmReleaseStatus(c *v1alpha1.PlatformV1alpha1Client, module *api.Module, deployedModule *api.Module) {
	status, err := helm.GetHelmReleaseStatus(deployedModule.GetName(), deployedModule.GetNamespace())
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	suite.gomega.Expect(status).To(gomega.Equal(helm.ReleaseStatusDeployed))
	suite.gomega.Expect(deployedModule.Status.LastSuccessfulVersion).To(gomega.Equal(module.Spec.Version))
}

func (suite *HelmModuleLifecycleTestSuite) verifyHelmValues(logger testLogger, c *v1alpha1.PlatformV1alpha1Client, module *api.Module, deployedModule *api.Module, overrides *apiextensionsv1.JSON, otherOverrides ...*map[string]interface{}) {
	suite.gomega.Eventually(func() bool {
		deployedValues, err := helm.GetValuesMap(vzlog.DefaultLogger(), deployedModule.GetName(), deployedModule.GetNamespace())
		if err != nil {
			t.Logf("error while fetching helm values from release %s/%s, %v", deployedModule.GetNamespace(), deployedModule.GetName(), err.Error())
			return false
		}

		appliedValuesBytes, err := overrides.MarshalJSON()
		if err != nil {
			t.Logf("unable to marshal override values, error: %v", err.Error())
			return false
		}

		var appliedValues map[string]interface{}
		err = json.Unmarshal(appliedValuesBytes, &appliedValues)
		if err != nil {
			logger.log("unable to unmarshal override values to map, error: %v", err.Error())
			return false
		}

		for _, otherOverride := range otherOverrides {
			err = yaml.MergeMaps(appliedValues, *otherOverride)
			if err != nil {
				logger.log("unable to merge override values, error: %v", err.Error())
				return false
			}
		}

		return reflect.DeepEqual(appliedValues, deployedValues)
	}, shortWaitTimeout, shortPollingInterval).Should(gomega.BeTrue())
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

func (suite *HelmModuleLifecycleTestSuite) verifyModuleDeleted(t *testing.T, c *v1alpha1.PlatformV1alpha1Client, module *api.Module) {
	suite.waitForModuleToBeDeleted(t, c, module.GetNamespace(), module.GetName())
	helmReleaseInstalled, err := helm.IsReleaseInstalled(module.GetName(), module.GetNamespace())
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	suite.gomega.Expect(helmReleaseInstalled).To(gomega.BeFalse())
}

func (suite *HelmModuleLifecycleTestSuite) deleteNamespace(ns string) bool {
	return ns != "default"
}

func (suite *HelmModuleLifecycleTestSuite) waitForModuleToBeReady(t *testing.T, c *v1alpha1.PlatformV1alpha1Client, module *api.Module) *api.Module {
	var deployedModule *api.Module
	var err error
	suite.gomega.Eventually(func() (bool, error) {
		deployedModule, err = c.Modules(module.GetNamespace()).Get(context.TODO(), module.GetName(), v1.GetOptions{})
		if err != nil {
			logger.log("error while fetching module %s/%s, %v", module.GetNamespace(), module.GetName(), err.Error())
			return false, err
		}

		cond := status.GetReadyCondition(deployedModule)
		if cond == nil {
			return false, nil
		}
		return cond.Status == corev1.ConditionTrue, nil
	}, shortWaitTimeout, shortPollingInterval).Should(gomega.BeTrue())
	return deployedModule
}

func (suite *HelmModuleLifecycleTestSuite) waitForModuleToBeDeleted(t *testing.T, c *v1alpha1.PlatformV1alpha1Client, namespace string, name string) {
	suite.gomega.Eventually(func() bool {
		_, err := c.Modules(namespace).Get(context.TODO(), name, v1.GetOptions{})
		if err != nil {
			if kerrs.IsNotFound(err) {
				return true
			}
			t.Logf("error while fetching module %s/%s, %v", namespace, name, err.Error())
			return false
		}
		return false
	}, shortWaitTimeout, shortPollingInterval).Should(gomega.BeTrue())
}

func (suite *HelmModuleLifecycleTestSuite) waitForNamespaceDeleted(t *testing.T, namespace string) error {
	c, err := k8sutil.GetCoreV1Client()
	if err != nil {
		return err
	}

	suite.gomega.Eventually(func() bool {
		_, err := c.Namespaces().Get(context.TODO(), namespace, v1.GetOptions{})
		if err != nil {
			if kerrs.IsNotFound(err) {
				return true
			}
			t.Logf("error while fetching namespace %s, %v", namespace, err.Error())
			return false
		}

		return false
	}, shortWaitTimeout, shortPollingInterval).Should(gomega.BeTrue())
	return err
}

func (suite *HelmModuleLifecycleTestSuite) waitForNamespaceCreated(t *testing.T, namespace string) error {
	c, err := k8sutil.GetCoreV1Client()
	if err != nil {
		return err
	}

	suite.gomega.Eventually(func() bool {
		_, err := c.Namespaces().Get(context.TODO(), namespace, v1.GetOptions{})
		if err != nil {
			if kerrs.IsNotFound(err) {
				_, err = c.Namespaces().Create(context.TODO(), &corev1.Namespace{ObjectMeta: v1.ObjectMeta{Name: namespace}}, v1.CreateOptions{})
				if err != nil {
					t.Logf("error while creating namespace %s, %v", namespace, err.Error())
				}

				return false
			}
			t.Logf("error while fetching namespace %s, %v", namespace, err.Error())
			return false
		}

		return true
	}, shortWaitTimeout, shortPollingInterval).Should(gomega.BeTrue())
	return err
}

func (suite *HelmModuleLifecycleTestSuite) waitForModuleToBeUpgraded(t *testing.T, c *v1alpha1.PlatformV1alpha1Client, module *api.Module) *api.Module {
	var deployedModule *api.Module
	var err error
	suite.gomega.Eventually(func() bool {
		deployedModule, err = c.Modules(module.GetNamespace()).Get(context.TODO(), module.GetName(), v1.GetOptions{})
		if err != nil {
			t.Logf("error while fetching module %s/%s, %v", module.GetNamespace(), module.GetName(), err.Error())
			return false
		}

		cond := status.GetReadyCondition(deployedModule)
		if cond == nil {
			return false
		}
		return cond.Status == corev1.ConditionTrue && cond.Reason == api.ReadyReasonUpgradeSucceeded &&
			deployedModule.Status.LastSuccessfulVersion == module.Spec.Version
	}, shortWaitTimeout, shortPollingInterval).Should(gomega.BeTrue())
	return deployedModule
}

func (suite *HelmModuleLifecycleTestSuite) removeModuleAndNamespace(logger testLogger, c *v1alpha1.PlatformV1alpha1Client, module *api.Module) {
	logger.log("delete module %s, version %s, namespace %s", module.GetName(), module.Spec.Version, module.GetNamespace())
	err := c.Modules(module.GetNamespace()).Delete(context.TODO(), module.GetName(), v1.DeleteOptions{})
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	suite.verifyModuleDeleted(logger, c, module)
	if suite.deleteNamespace(module.GetNamespace()) {
		corev1client, err := k8sutil.GetCoreV1Client()
		suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
		err = corev1client.Namespaces().Delete(context.TODO(), module.GetNamespace(), v1.DeleteOptions{})
		suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}
}
