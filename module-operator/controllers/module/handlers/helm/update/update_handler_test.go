// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package update

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/time"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/handlers/common"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/handlerspi"
	vzhelm "github.com/verrazzano/verrazzano-modules/pkg/helm"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
)

const (
	releaseName = "test-release"
	namespace   = "test-ns"
	moduleName  = "test-module"
)

// TestGetWorkName tests the update handler GetWorkName function
func TestGetWorkName(t *testing.T) {
	asserts := assert.New(t)
	handler := NewHandler()

	// GIVEN an update handler
	// WHEN the GetWorkName function is called
	// THEN it returns the expected work name
	workName := handler.GetWorkName()
	asserts.Equal("update", workName)
}

// TestPreWorkUpdateStatus tests the update handler PreWorkUpdateStatus function
func TestPreWorkUpdateStatus(t *testing.T) {
	asserts := assert.New(t)
	handler := NewHandler()

	// GIVEN an update handler
	// WHEN the PreWorkUpdateStatus function is called
	// THEN no error occurs and the function returns an empty ctrl.Result and the Module status
	// has the expected state and condition
	module := &v1alpha1.Module{
		ObjectMeta: metav1.ObjectMeta{
			Name:      moduleName,
			Namespace: namespace,
		},
	}

	cli := fake.NewClientBuilder().WithScheme(newScheme()).WithObjects(module).Build()
	ctx := handlerspi.HandlerContext{
		Log:    vzlog.DefaultLogger(),
		Client: cli,
		CR:     module,
	}

	result, err := handler.PreWorkUpdateStatus(ctx)
	asserts.NoError(err)
	asserts.Equal(ctrl.Result{}, result)

	// fetch the Module and validate that the condition and state are set
	err = cli.Get(context.TODO(), types.NamespacedName{Name: moduleName, Namespace: namespace}, module)
	asserts.NoError(err)
	asserts.Equal(v1alpha1.CondPreInstall, module.Status.Conditions[0].Type)
	asserts.Equal(v1alpha1.ModuleStateReconciling, string(module.Status.State))
}

// TestPreWork tests the update handler PreWork function
func TestPreWork(t *testing.T) {
	asserts := assert.New(t)
	handler := NewHandler()

	// GIVEN an update handler
	// WHEN the PreWork function is called
	// THEN no error occurs and the function returns an empty ctrl.Result
	result, err := handler.PreWork(handlerspi.HandlerContext{})
	asserts.NoError(err)
	asserts.Equal(ctrl.Result{}, result)
}

// TestDoWorkUpdateStatus tests the update handler DoWorkUpdateStatus function
func TestDoWorkUpdateStatus(t *testing.T) {
	asserts := assert.New(t)
	handler := NewHandler()

	// GIVEN an update handler
	// WHEN the DoWorkUpdateStatus function is called
	// THEN no error occurs and the function returns an empty ctrl.Result and the Module status
	// has the expected state and condition
	module := &v1alpha1.Module{
		ObjectMeta: metav1.ObjectMeta{
			Name:      moduleName,
			Namespace: namespace,
		},
	}

	cli := fake.NewClientBuilder().WithScheme(newScheme()).WithObjects(module).Build()
	ctx := handlerspi.HandlerContext{
		Log:    vzlog.DefaultLogger(),
		Client: cli,
		CR:     module,
	}

	result, err := handler.DoWorkUpdateStatus(ctx)
	asserts.NoError(err)
	asserts.Equal(ctrl.Result{}, result)

	// fetch the Module and validate that the condition and state are set
	err = cli.Get(context.TODO(), types.NamespacedName{Name: moduleName, Namespace: namespace}, module)
	asserts.NoError(err)
	asserts.Equal(v1alpha1.CondInstallStarted, module.Status.Conditions[0].Type)
	asserts.Equal(v1alpha1.ModuleStateReconciling, string(module.Status.State))
}

func getChart() *chart.Chart {
	return &chart.Chart{
		Metadata: &chart.Metadata{
			APIVersion: "v1",
			Name:       "hello",
			Version:    "0.1.0",
			AppVersion: "1.0",
		},
		Templates: []*chart.File{
			{Name: "templates/hello", Data: []byte("hello: world")},
		},
	}
}

func createRelease(name string, status release.Status) *release.Release {
	now := time.Now()
	return &release.Release{
		Name:      releaseName,
		Namespace: namespace,
		Info: &release.Info{
			FirstDeployed: now,
			LastDeployed:  now,
			Status:        status,
			Description:   "Named Release Stub",
		},
		Chart: getChart(),
		Config: map[string]interface{}{
			"name1": "value1",
			"name2": "value2",
		},
		Version: 1,
	}
}

// testActionConfigWithRelease is a fake action that returns an installed Helm release
func testActionConfigWithRelease(log vzlog.VerrazzanoLogger, settings *cli.EnvSettings, namespace string) (*action.Configuration, error) {
	return vzhelm.CreateActionConfig(true, releaseName, release.StatusDeployed, log, createRelease)
}

// testActionConfigWithNoRelease is a fake action that returns an uninstalled Helm release
func testActionConfigWithNoRelease(log vzlog.VerrazzanoLogger, settings *cli.EnvSettings, namespace string) (*action.Configuration, error) {
	return vzhelm.CreateActionConfig(false, releaseName, release.StatusUninstalled, log, createRelease)
}

// TestDoWork tests the update handler DoWork function
func TestDoWork(t *testing.T) {
	asserts := assert.New(t)
	handler := NewHandler()

	// GIVEN an update handler
	// WHEN the DoWork function is called
	// THEN no error occurs, the function returns an empty ctrl.Result, and the Helm upgrade
	// function has been called
	cli := fake.NewClientBuilder().WithScheme(newScheme()).Build()
	ctx := handlerspi.HandlerContext{
		Log:    vzlog.DefaultLogger(),
		Client: cli,
		CR:     &v1alpha1.Module{},
		HelmInfo: handlerspi.HelmInfo{
			HelmRelease: &handlerspi.HelmRelease{
				Name:      releaseName,
				Namespace: namespace,
			},
		},
	}

	var upgradeFuncCalled = false
	common.SetUpgradeFunc(func(log vzlog.VerrazzanoLogger, releaseOpts *vzhelm.HelmReleaseOpts, wait bool, dryRun bool) (*release.Release, error) {
		upgradeFuncCalled = true
		return nil, nil
	})
	defer common.ResetUpgradeFunc()

	result, err := handler.DoWork(ctx)
	asserts.NoError(err)
	asserts.Equal(ctrl.Result{}, result)
	asserts.True(upgradeFuncCalled)
}

// TestIsWorkDone tests the update handler IsWorkDone function
func TestIsWorkDone(t *testing.T) {
	asserts := assert.New(t)

	vzhelm.SetActionConfigFunction(testActionConfigWithRelease)
	defer vzhelm.SetDefaultActionConfigFunction()

	handler := NewHandler()

	// GIVEN an update handler and no deployed workloads
	// WHEN the IsWorkDone function is called
	// THEN no error occurs and the function returns true and an empty ctrl.Result
	cli := fake.NewClientBuilder().WithScheme(newScheme()).Build()
	ctx := handlerspi.HandlerContext{
		Log:    vzlog.DefaultLogger(),
		Client: cli,
		CR:     &v1alpha1.Module{},
		HelmInfo: handlerspi.HelmInfo{
			HelmRelease: &handlerspi.HelmRelease{
				Name:      releaseName,
				Namespace: namespace,
			},
		},
	}

	done, result, err := handler.IsWorkDone(ctx)
	asserts.NoError(err)
	asserts.True(done)
	asserts.Equal(ctrl.Result{}, result)
}

// TestIsWorkNeeded tests the update handler IsWorkNeeded function
func TestIsWorkNeeded(t *testing.T) {
	asserts := assert.New(t)

	vzhelm.SetActionConfigFunction(testActionConfigWithRelease)
	defer vzhelm.SetDefaultActionConfigFunction()

	handler := NewHandler()

	// GIVEN an update handler and a Helm release that is already installed
	// WHEN the IsWorkNeeded function is called
	// THEN no error occurs and the function returns true and an empty ctrl.Result
	cli := fake.NewClientBuilder().WithScheme(newScheme()).Build()
	ctx := handlerspi.HandlerContext{
		Log:    vzlog.DefaultLogger(),
		Client: cli,
		CR:     &v1alpha1.Module{},
		HelmInfo: handlerspi.HelmInfo{
			HelmRelease: &handlerspi.HelmRelease{
				Name:      releaseName,
				Namespace: namespace,
			},
		},
	}

	needed, result, err := handler.IsWorkNeeded(ctx)
	asserts.NoError(err)
	asserts.True(needed)
	asserts.Equal(ctrl.Result{}, result)

	// GIVEN an update handler and a Helm release that is not installed
	// WHEN the IsWorkNeeded function is called
	// THEN no error occurs and the function returns false and an empty ctrl.Result
	vzhelm.SetActionConfigFunction(testActionConfigWithNoRelease)

	needed, result, err = handler.IsWorkNeeded(ctx)
	asserts.NoError(err)
	asserts.False(needed)
	asserts.Equal(ctrl.Result{}, result)
}

// TestPostWorkUpdateStatus tests the update handler PostWorkUpdateStatus function
func TestPostWorkUpdateStatus(t *testing.T) {
	asserts := assert.New(t)

	handler := NewHandler()

	// GIVEN an update handler
	// WHEN the PostWorkUpdateStatus function is called
	// THEN no error occurs and the function returns an empty ctrl.Result
	result, err := handler.PostWorkUpdateStatus(handlerspi.HandlerContext{})
	asserts.NoError(err)
	asserts.Equal(ctrl.Result{}, result)
}

// TestPostWork tests the update handler PostWork function
func TestPostWork(t *testing.T) {
	asserts := assert.New(t)

	handler := NewHandler()

	// GIVEN an update handler
	// WHEN the PostWork function is called
	// THEN no error occurs and the function returns an empty ctrl.Result
	result, err := handler.PostWork(handlerspi.HandlerContext{})
	asserts.NoError(err)
	asserts.Equal(ctrl.Result{}, result)
}

// TestWorkCompletedUpdateStatus tests the update handler WorkCompletedUpdateStatus function
func TestWorkCompletedUpdateStatus(t *testing.T) {
	asserts := assert.New(t)
	handler := NewHandler()

	// GIVEN an update handler
	// WHEN the WorkCompletedUpdateStatus function is called
	// THEN no error occurs and the function returns an empty ctrl.Result and the Module status
	// has the expected state and condition
	module := &v1alpha1.Module{
		ObjectMeta: metav1.ObjectMeta{
			Name:      moduleName,
			Namespace: namespace,
		},
	}

	cli := fake.NewClientBuilder().WithScheme(newScheme()).WithObjects(module).Build()
	ctx := handlerspi.HandlerContext{
		Log:    vzlog.DefaultLogger(),
		Client: cli,
		CR:     module,
	}

	result, err := handler.WorkCompletedUpdateStatus(ctx)
	asserts.NoError(err)
	asserts.Equal(ctrl.Result{}, result)

	// fetch the Module and validate that the condition and state are set
	err = cli.Get(context.TODO(), types.NamespacedName{Name: moduleName, Namespace: namespace}, module)
	asserts.NoError(err)
	asserts.Equal(v1alpha1.CondInstallComplete, module.Status.Conditions[0].Type)
	asserts.Equal(v1alpha1.ModuleStateReady, string(module.Status.State))
}

func newScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)
	return scheme
}
