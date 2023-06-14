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
	"github.com/stretchr/testify/suite"
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
	suite.Run(t, helmModuleLifecyclreTestingSuite)
}

func (suite *HelmModuleLifecycleTestSuite) executeModuleLifecycleOperationsWithTargetNS(ctx *common.TestContext, namespace, targetNamespace string) {
	suite.waitForNamespaceCreated(ctx, namespace)
	module := &api.Module{}
	err := common.UnmarshalTestFile(common.TEST_HELM_MODULE_FILE, module)
	ctx.GomegaWithT.Expect(err).NotTo(gomega.HaveOccurred())

	module.SetNamespace(namespace)
	module.Spec.TargetNamespace = targetNamespace

	testNamespacesMutex.Lock()
	testNamespaces[namespace] = []string{module.GetName()}
	testNamespaces[module.Spec.TargetNamespace] = []string{module.GetName()}
	testNamespacesMutex.Unlock()

	module, overrides := suite.createOrUpdateModule(ctx, module, common.TEST_HELM_MODULE_OVERRIDE_010, false)
	module = suite.verifyModule(ctx, module, overrides)

	module, overrides = suite.createOrUpdateModule(ctx, module, common.TEST_HELM_MODULE_OVERRIDE_010_1, true)
	module = suite.verifyModule(ctx, module, overrides)

	module.Spec.Version = common.TEST_HELM_MODULE_VERSION_011
	module, overrides = suite.createOrUpdateModule(ctx, module, common.TEST_HELM_MODULE_OVERRIDE_011, true)
	module = suite.waitForModuleToBeUpgraded(ctx, module)
	module = suite.verifyModule(ctx, module, overrides)
	suite.removeModuleAndNamespace(ctx, module)
}

func (suite *HelmModuleLifecycleTestSuite) cleanup() {
	ctx := common.NewTestContext(suite.T())
	for namespace, modules := range testNamespaces {
		if suite.deleteNamespace(namespace) {
			err := ctx.CoreV1Client().Namespaces().Delete(context.TODO(), namespace, v1.DeleteOptions{})
			if kerrs.IsNotFound(err) {
				continue
			}

			ctx.GomegaWithT.Expect(err).NotTo(gomega.HaveOccurred())
			suite.waitForNamespaceDeleted(ctx, namespace)
			ctx.GomegaWithT.Expect(err).NotTo(gomega.HaveOccurred())
		} else {
			for _, moduleName := range modules {
				err := ctx.ModuleClient().Modules(namespace).Delete(context.TODO(), moduleName, v1.DeleteOptions{})
				if kerrs.IsNotFound(err) {
					continue
				}

				ctx.GomegaWithT.Expect(err).NotTo(gomega.HaveOccurred())
				suite.waitForModuleToBeDeleted(ctx, namespace, moduleName)
			}
		}

	}
}

func (suite *HelmModuleLifecycleTestSuite) createOrUpdateModule(ctx *common.TestContext, module *api.Module, overridesFile string, update bool, otherOverrides ...*api.ValuesFromSource) (*api.Module, *apiextensionsv1.JSON) {
	op := "create"
	if update {
		op = "update"
	}
	name := module.Name
	namespace := module.Namespace
	version := module.Spec.Version

	ctx.T.Logf("%s module %s, version %s, namespace %s", op, module.GetName(), module.Spec.Version, module.GetNamespace())

	// Build the values and valuesFrom
	values := suite.generateOverridesFromFile(ctx, overridesFile)
	var valuesFrom []api.ValuesFromSource
	for _, toAppend := range otherOverrides {
		valuesFrom = append(valuesFrom, *toAppend)
	}
	ctx.GomegaWithT.Eventually(func() error {
		var err error

		// Get the latest Module or else the code will never resolve conflicts
		if op != "create" {
			module, err = ctx.ModuleClient().Modules(namespace).Get(context.TODO(), name, v1.GetOptions{})
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
			module, err = ctx.ModuleClient().Modules(module.GetNamespace()).Update(context.TODO(), module, v1.UpdateOptions{})
		} else {
			module, err = ctx.ModuleClient().Modules(module.GetNamespace()).Create(context.TODO(), module, v1.CreateOptions{})
		}

		return err
	}, shortWaitTimeout, shortPollingInterval).ShouldNot(gomega.HaveOccurred())

	return module, values
}

func (suite *HelmModuleLifecycleTestSuite) verifyModule(ctx *common.TestContext, module *api.Module, overrides *apiextensionsv1.JSON, otherOverrides ...*map[string]interface{}) *api.Module {
	deployedModule := suite.verifyModuleIsReady(ctx, module)
	suite.verifyHelmReleaseStatus(ctx, module, deployedModule)
	suite.verifyHelmValues(ctx, module, deployedModule, overrides, otherOverrides...)
	deployedModule, err := ctx.ModuleClient().Modules(module.GetNamespace()).Get(context.TODO(), module.GetName(), v1.GetOptions{})
	ctx.GomegaWithT.Expect(err).NotTo(gomega.HaveOccurred())
	return deployedModule
}

func (suite *HelmModuleLifecycleTestSuite) verifyModuleIsReady(ctx *common.TestContext, module *api.Module) *api.Module {
	deployedModule := suite.waitForModuleToBeReady(ctx, module)
	ctx.GomegaWithT.Expect(deployedModule.Status.LastSuccessfulVersion).To(gomega.Equal(module.Spec.Version))
	return deployedModule
}

func (suite *HelmModuleLifecycleTestSuite) verifyHelmReleaseStatus(ctx *common.TestContext, module *api.Module, deployedModule *api.Module) {
	status, err := helm.GetHelmReleaseStatus(deployedModule.GetName(), deployedModule.Spec.TargetNamespace)
	ctx.GomegaWithT.Expect(err).NotTo(gomega.HaveOccurred())
	ctx.GomegaWithT.Expect(status).To(gomega.Equal(helm.ReleaseStatusDeployed))
	ctx.GomegaWithT.Expect(deployedModule.Status.LastSuccessfulVersion).To(gomega.Equal(module.Spec.Version))
}

func (suite *HelmModuleLifecycleTestSuite) verifyHelmValues(ctx *common.TestContext, module *api.Module, deployedModule *api.Module, overrides *apiextensionsv1.JSON, otherOverrides ...*map[string]interface{}) {
	ctx.GomegaWithT.Eventually(func() bool {
		deployedValues, err := helm.GetValuesMap(vzlog.DefaultLogger(), deployedModule.GetName(), deployedModule.Spec.TargetNamespace)
		if err != nil {
			ctx.T.Logf("error while fetching helm values from release %s/%s, %v", deployedModule.Spec.TargetNamespace, deployedModule.GetName(), err.Error())
			return false
		}

		appliedValuesBytes, err := overrides.MarshalJSON()
		if err != nil {
			ctx.T.Logf("unable to marshal override values, error: %v", err.Error())
			return false
		}

		var appliedValues map[string]interface{}
		err = json.Unmarshal(appliedValuesBytes, &appliedValues)
		if err != nil {
			ctx.T.Logf("unable to unmarshal override values to map, error: %v", err.Error())
			return false
		}

		for _, otherOverride := range otherOverrides {
			err = yaml.MergeMaps(appliedValues, *otherOverride)
			if err != nil {
				ctx.T.Logf("unable to merge override values, error: %v", err.Error())
				return false
			}
		}

		return reflect.DeepEqual(appliedValues, deployedValues)
	}, shortWaitTimeout, shortPollingInterval).Should(gomega.BeTrue())
}

func (suite *HelmModuleLifecycleTestSuite) generateOverridesFromFile(ctx *common.TestContext, overridesFile string) *apiextensionsv1.JSON {
	overrides := &apiextensionsv1.JSON{}
	err := common.UnmarshalTestFile(overridesFile, overrides)
	ctx.GomegaWithT.Expect(err).NotTo(gomega.HaveOccurred())
	return overrides
}

func (suite *HelmModuleLifecycleTestSuite) verifyModuleDeleted(ctx *common.TestContext, module *api.Module) {
	suite.waitForModuleToBeDeleted(ctx, module.GetNamespace(), module.GetName())
	helmReleaseInstalled, err := helm.IsReleaseInstalled(module.GetName(), module.Spec.TargetNamespace)
	ctx.GomegaWithT.Expect(err).NotTo(gomega.HaveOccurred())
	ctx.GomegaWithT.Expect(helmReleaseInstalled).To(gomega.BeFalse())
}

func (suite *HelmModuleLifecycleTestSuite) deleteNamespace(ns string) bool {
	return ns != "default"
}

func (suite *HelmModuleLifecycleTestSuite) waitForModuleToBeReady(ctx *common.TestContext, module *api.Module) *api.Module {
	var deployedModule *api.Module
	var err error
	ctx.GomegaWithT.Eventually(func() (bool, error) {
		deployedModule, err = ctx.ModuleClient().Modules(module.GetNamespace()).Get(context.TODO(), module.GetName(), v1.GetOptions{})
		if err != nil {
			ctx.T.Logf("error while fetching module %s/%s, %v", module.GetNamespace(), module.GetName(), err.Error())
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

func (suite *HelmModuleLifecycleTestSuite) waitForModuleToBeDeleted(ctx *common.TestContext, namespace string, name string) {
	ctx.GomegaWithT.Eventually(func() bool {
		_, err := ctx.ModuleClient().Modules(namespace).Get(context.TODO(), name, v1.GetOptions{})
		if err != nil {
			if kerrs.IsNotFound(err) {
				return true
			}
			ctx.T.Logf("error while fetching module %s/%s, %v", namespace, name, err.Error())
			return false
		}
		return false
	}, shortWaitTimeout, shortPollingInterval).Should(gomega.BeTrue())
}

func (suite *HelmModuleLifecycleTestSuite) waitForNamespaceDeleted(ctx *common.TestContext, namespace string) {
	ctx.GomegaWithT.Eventually(func() bool {
		_, err := ctx.CoreV1Client().Namespaces().Get(context.TODO(), namespace, v1.GetOptions{})
		if err != nil {
			if kerrs.IsNotFound(err) {
				return true
			}
			ctx.T.Logf("error while fetching namespace %s, %v", namespace, err.Error())
			return false
		}

		return false
	}, shortWaitTimeout, shortPollingInterval).Should(gomega.BeTrue())
}

func (suite *HelmModuleLifecycleTestSuite) waitForNamespaceCreated(ctx *common.TestContext, namespace string) {
	ctx.GomegaWithT.Eventually(func() bool {
		_, err := ctx.CoreV1Client().Namespaces().Get(context.TODO(), namespace, v1.GetOptions{})
		if err != nil {
			if kerrs.IsNotFound(err) {
				_, err = ctx.CoreV1Client().Namespaces().Create(context.TODO(), &corev1.Namespace{ObjectMeta: v1.ObjectMeta{Name: namespace}}, v1.CreateOptions{})
				if err != nil {
					ctx.T.Logf("error while creating namespace %s, %v", namespace, err.Error())
				}

				return false
			}
			ctx.T.Logf("error while fetching namespace %s, %v", namespace, err.Error())
			return false
		}

		return true
	}, shortWaitTimeout, shortPollingInterval).Should(gomega.BeTrue())
}

func (suite *HelmModuleLifecycleTestSuite) waitForModuleToBeUpgraded(ctx *common.TestContext, module *api.Module) *api.Module {
	var deployedModule *api.Module
	var err error
	ctx.GomegaWithT.Eventually(func() bool {
		deployedModule, err = ctx.ModuleClient().Modules(module.GetNamespace()).Get(context.TODO(), module.GetName(), v1.GetOptions{})
		if err != nil {
			ctx.T.Logf("error while fetching module %s/%s, %v", module.GetNamespace(), module.GetName(), err.Error())
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

func (suite *HelmModuleLifecycleTestSuite) removeModuleAndNamespace(ctx *common.TestContext, module *api.Module) {
	ctx.T.Logf("delete module %s, version %s, namespace %s", module.GetName(), module.Spec.Version, module.GetNamespace())
	err := ctx.ModuleClient().Modules(module.GetNamespace()).Delete(context.TODO(), module.GetName(), v1.DeleteOptions{})
	ctx.GomegaWithT.Expect(err).NotTo(gomega.HaveOccurred())
	suite.verifyModuleDeleted(ctx, module)
	if suite.deleteNamespace(module.GetNamespace()) {
		corev1client, err := k8sutil.GetCoreV1Client()
		ctx.GomegaWithT.Expect(err).NotTo(gomega.HaveOccurred())
		err = corev1client.Namespaces().Delete(context.TODO(), module.GetNamespace(), v1.DeleteOptions{})
		ctx.GomegaWithT.Expect(err).NotTo(gomega.HaveOccurred())
	}
}

func (suite *HelmModuleLifecycleTestSuite) executeModuleLifecycleOperations(ctx *common.TestContext, namespace string) {
	suite.executeModuleLifecycleOperationsWithTargetNS(ctx, namespace, namespace)
}
