// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package helm_module_lifecycle_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/status"
	"strings"
	"sync"
	"testing"
	"time"

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

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	kerrs "k8s.io/apimachinery/pkg/api/errors"
)

// HelmModuleLifecycleTestSuite is the test suite for the lifecycle tests for the helm module.
type HelmModuleLifecycleTestSuite struct {
	suite.Suite
	gomega types.Gomega
}

type testLogger struct {
	testName string
}

func (logger *testLogger) log(format string, args ...any) {
	fmt.Printf(fmt.Sprintf("%s: ", logger.testName)+format+"\n", args...)
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

func (suite *HelmModuleLifecycleTestSuite) executeModuleLifecycleOperations(testName string, namespace string) {
	logger := testLogger{testName: testName}
	err := suite.waitForNamespaceCreated(logger, namespace)
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
	module, overrides := suite.createOrUpdateModule(logger, c, module, common.TEST_HELM_MODULE_OVERRIDE_010, false)
	module = suite.verifyModule(logger, c, module, overrides)

	module, overrides = suite.createOrUpdateModule(logger, c, module, common.TEST_HELM_MODULE_OVERRIDE_010_1, true)
	module = suite.verifyModule(logger, c, module, overrides)

	module.Spec.Version = common.TEST_HELM_MODULE_VERSION_011
	module, overrides = suite.createOrUpdateModule(logger, c, module, common.TEST_HELM_MODULE_OVERRIDE_011, true)
	module = suite.waitForModuleToBeUpgraded(logger, c, module)
	module = suite.verifyModule(logger, c, module, overrides)
	logger.log("delete module %s, version %s, namespace %s", module.GetName(), module.Spec.Version, module.GetNamespace())
	err = c.Modules(module.GetNamespace()).Delete(context.TODO(), module.GetName(), v1.DeleteOptions{})
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	suite.verifyModuleDeleted(logger, c, module)
	if suite.deleteNamespace(namespace) {
		corev1client, err := k8sutil.GetCoreV1Client()
		suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
		err = corev1client.Namespaces().Delete(context.TODO(), namespace, v1.DeleteOptions{})
		suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}

}

func (suite *HelmModuleLifecycleTestSuite) cleanup() {
	logger := testLogger{testName: suite.T().Name()}
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
			err = suite.waitForNamespaceDeleted(logger, namespace)
			suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
		} else {
			for _, moduleName := range modules {
				err = c.Modules(namespace).Delete(context.TODO(), moduleName, v1.DeleteOptions{})
				if kerrs.IsNotFound(err) {
					continue
				}

				suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
				suite.waitForModuleToBeDeleted(logger, c, namespace, moduleName)
			}
		}

	}
}

func (suite *HelmModuleLifecycleTestSuite) createOrUpdateModule(logger testLogger, c *v1alpha1.PlatformV1alpha1Client, module *api.Module, overridesFile string, update bool) (*api.Module, *apiextensionsv1.JSON) {
	var err error
	op := "create"
	if update {
		op = "update"
		module, err = c.Modules(module.GetNamespace()).Get(context.TODO(), module.GetName(), v1.GetOptions{})
		suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}

	logger.log("%s module %s, version %s, namespace %s", op, module.GetName(), module.Spec.Version, module.GetNamespace())
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

func (suite *HelmModuleLifecycleTestSuite) verifyModule(logger testLogger, c *v1alpha1.PlatformV1alpha1Client, module *api.Module, overrides *apiextensionsv1.JSON) *api.Module {
	deployedModule := suite.verifyModuleIsReady(logger, c, module)
	suite.verifyHelmReleaseStatus(c, module, deployedModule)
	suite.verifyHelmValues(logger, c, module, deployedModule, overrides)
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

func (suite *HelmModuleLifecycleTestSuite) verifyHelmValues(logger testLogger, c *v1alpha1.PlatformV1alpha1Client, module *api.Module, deployedModule *api.Module, overrides *apiextensionsv1.JSON) {
	suite.gomega.Eventually(func() bool {
		deployedValues, err := helm.GetValuesMap(vzlog.DefaultLogger(), deployedModule.GetName(), deployedModule.GetNamespace())
		if err != nil {
			logger.log("error while fetching helm values from release %s/%s, %v", deployedModule.GetNamespace(), deployedModule.GetName(), err.Error())
			return false
		}

		deployedValuesBytes, err := json.Marshal(deployedValues)
		if err != nil {
			logger.log("unable to marshal helm values, error: %v", err.Error())
			return false
		}

		appliedValuesBytes, err := overrides.MarshalJSON()
		if err != nil {
			logger.log("unable to marshal override values, error: %v", err.Error())
			return false
		}

		return bytes.Equal(deployedValuesBytes, appliedValuesBytes)
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

func (suite *HelmModuleLifecycleTestSuite) verifyModuleDeleted(logger testLogger, c *v1alpha1.PlatformV1alpha1Client, module *api.Module) {
	suite.waitForModuleToBeDeleted(logger, c, module.GetNamespace(), module.GetName())
	helmReleaseInstalled, err := helm.IsReleaseInstalled(module.GetName(), module.GetNamespace())
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	suite.gomega.Expect(helmReleaseInstalled).To(gomega.BeFalse())
}

func (suite *HelmModuleLifecycleTestSuite) deleteNamespace(ns string) bool {
	return strings.HasPrefix(ns, "ns")
}

func (suite *HelmModuleLifecycleTestSuite) waitForModuleToBeReady(logger testLogger, c *v1alpha1.PlatformV1alpha1Client, module *api.Module) *api.Module {
	var deployedModule *api.Module
	var err error
	suite.gomega.Eventually(func() bool {
		deployedModule, err = c.Modules(module.GetNamespace()).Get(context.TODO(), module.GetName(), v1.GetOptions{})
		if err != nil {
			logger.log("error while fetching module %s/%s, %v", module.GetNamespace(), module.GetName(), err.Error())
			return false
		}

		cond := status.GetReadyCondition(module)
		if cond == nil {
			return false
		}
		return cond.Status == corev1.ConditionTrue
	}, shortWaitTimeout, shortPollingInterval).Should(gomega.BeTrue())
	return deployedModule
}

func (suite *HelmModuleLifecycleTestSuite) waitForModuleToBeDeleted(logger testLogger, c *v1alpha1.PlatformV1alpha1Client, namespace string, name string) {
	suite.gomega.Eventually(func() bool {
		_, err := c.Modules(namespace).Get(context.TODO(), name, v1.GetOptions{})
		if err != nil {
			if kerrs.IsNotFound(err) {
				return true
			}
			logger.log("error while fetching module %s/%s, %v", namespace, name, err.Error())
			return false
		}
		return false
	}, shortWaitTimeout, shortPollingInterval).Should(gomega.BeTrue())
}

func (suite *HelmModuleLifecycleTestSuite) waitForNamespaceDeleted(logger testLogger, namespace string) error {
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
			logger.log("error while fetching namespace %s, %v", namespace, err.Error())
			return false
		}

		return false
	}, shortWaitTimeout, shortPollingInterval).Should(gomega.BeTrue())
	return err
}

func (suite *HelmModuleLifecycleTestSuite) waitForNamespaceCreated(logger testLogger, namespace string) error {
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
					logger.log("error while creating namespace %s, %v", namespace, err.Error())
				}

				return false
			}
			logger.log("error while fetching namespace %s, %v", namespace, err.Error())
			return false
		}

		return true
	}, shortWaitTimeout, shortPollingInterval).Should(gomega.BeTrue())
	return err
}

func (suite *HelmModuleLifecycleTestSuite) waitForModuleToBeUpgraded(logger testLogger, c *v1alpha1.PlatformV1alpha1Client, module *api.Module) *api.Module {
	var deployedModule *api.Module
	var err error
	suite.gomega.Eventually(func() bool {
		deployedModule, err = c.Modules(module.GetNamespace()).Get(context.TODO(), module.GetName(), v1.GetOptions{})
		if err != nil {
			logger.log("error while fetching module %s/%s, %v", module.GetNamespace(), module.GetName(), err.Error())
			return false
		}

		cond := status.GetReadyCondition(module)
		if cond == nil {
			return false
		}
		return cond.Status == corev1.ConditionTrue &&
			cond.Reason == api.ReadyReasonUpgradeSucceeded &&
			deployedModule.Status.LastSuccessfulVersion == module.Spec.Version
	}, shortWaitTimeout, shortPollingInterval).Should(gomega.BeTrue())
	return deployedModule
}
