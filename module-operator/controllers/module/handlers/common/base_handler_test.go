// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package common

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/handlerspi"
	vzhelm "github.com/verrazzano/verrazzano-modules/pkg/helm"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/time"

	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakes "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

const (
	releaseName      = "release"
	releaseNamespace = "releaseNS"
	namespace        = "test-ns"
	moduleName       = "test-module"
)

type fakeHandler struct {
	BaseHandler
	*vzhelm.HelmReleaseOpts
	err   error
	ready bool
}

// TestHelmUpgradeOrInstall tests the Helm upgrade and install
// GIVEN a chart and release information
// WHEN HelmUpgradeOrInstall is called
// THEN ensure that correct parameters are passed to the upgradeFunc
func TestHelmUpgradeOrInstall(t *testing.T) {
	asserts := assert.New(t)
	tests := []struct {
		name             string
		releaseName      string
		releaseNamespace string
		chartPath        string
		chartVersion     string
		repoURL          string
		err              error
	}{
		{
			name:             "test-success",
			releaseName:      "rel1",
			releaseNamespace: "testns",
			chartPath:        "testpath",
			chartVersion:     "v1.0",
			repoURL:          "url",
		},
		{
			name:             "test-err",
			releaseName:      "rel1",
			releaseNamespace: "testns",
			chartPath:        "testpath",
			chartVersion:     "v1.0",
			repoURL:          "url",
			err:              fmt.Errorf("fake-error"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cli := fakes.NewClientBuilder().WithScheme(newScheme()).WithObjects().Build()
			module := &moduleapi.Module{
				ObjectMeta: metav1.ObjectMeta{
					Name:      moduleName,
					Namespace: namespace,
				},
			}

			rctx := handlerspi.HandlerContext{
				Client: cli,
				Log:    vzlog.DefaultLogger(),
				DryRun: false,
				CR:     module,
				HelmInfo: handlerspi.HelmInfo{
					HelmRelease: &handlerspi.HelmRelease{
						Name:      test.releaseName,
						Namespace: test.releaseNamespace,
						ChartInfo: handlerspi.HelmChart{
							Version: test.chartVersion,
							Path:    test.chartPath,
						},
						Repository: handlerspi.HelmChartRepository{
							URI: test.repoURL,
						},
					},
				},
			}
			defer ResetUpgradeFunc()
			h := fakeHandler{err: test.err}
			upgradeFunc = h.upgradeFunc

			result := h.HelmUpgradeOrInstall(rctx)
			asserts.Equal(test.err, result.GetError())
			asserts.Equal(test.chartPath, h.ChartPath)
			asserts.Equal(test.chartVersion, h.ChartVersion)
			asserts.Equal(test.repoURL, h.RepoURL)
			asserts.Equal(test.releaseNamespace, h.Namespace)
			asserts.Equal(test.releaseName, h.ReleaseName)
		})
	}
}

// TestCheckReleaseDeployedAndReady tests the Helm release is deployed and ready
// GIVEN a Helm release
// WHEN CheckReleaseDeployedAndReady is called
// THEN ensure that correct result is returned
func TestCheckReleaseDeployedAndReady(t *testing.T) {
	vzhelm.SetActionConfigFunction(testActionConfigWithRelease)
	defer vzhelm.SetDefaultActionConfigFunction()

	asserts := assert.New(t)
	tests := []struct {
		name             string
		releaseName      string
		releaseNamespace string
		err              error
		ready            bool
	}{
		{
			name:             "test-ready",
			releaseName:      releaseName,
			releaseNamespace: releaseNamespace,
			ready:            true,
		},
		{
			name:             "test-not-ready",
			releaseName:      releaseName,
			releaseNamespace: releaseNamespace,
			ready:            false,
		},
		{
			name:             "test-err",
			releaseName:      releaseName,
			releaseNamespace: releaseNamespace,
			ready:            false,
			err:              fmt.Errorf("fake-err"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cli := fakes.NewClientBuilder().WithScheme(newScheme()).WithObjects().Build()
			rctx := handlerspi.HandlerContext{
				Client: cli,
				Log:    vzlog.DefaultLogger(),
				DryRun: false,
				HelmInfo: handlerspi.HelmInfo{
					HelmRelease: &handlerspi.HelmRelease{
						Name:      test.releaseName,
						Namespace: test.releaseNamespace,
					},
				},
			}
			defer ResetCheckReadyFunc()
			h := fakeHandler{err: test.err, ready: test.ready}
			SetCheckReadyFunc(h.checkWorkLoadsReady)

			ready, result := h.CheckReleaseDeployedAndReady(rctx)
			asserts.Equal(test.err, result.GetError())
			asserts.Equal(test.ready, ready)
		})
	}
}

func (f *fakeHandler) upgradeFunc(log vzlog.VerrazzanoLogger, releaseOpts *vzhelm.HelmReleaseOpts, wait bool, dryRun bool) (*release.Release, error) {
	f.HelmReleaseOpts = releaseOpts
	return nil, f.err
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

func (f *fakeHandler) checkWorkLoadsReady(ctx handlerspi.HandlerContext, releaseName string, namespace string) (bool, error) {
	return f.ready, f.err

}

// testActionConfigWithRelease is a fake action that returns an installed Helm release
func testActionConfigWithRelease(log vzlog.VerrazzanoLogger, settings *cli.EnvSettings, namespace string) (*action.Configuration, error) {
	return vzhelm.CreateActionConfig(true, releaseName, release.StatusDeployed, log, createRelease)
}

// testActionConfigWithNoRelease is a fake action that returns an uninstalled Helm release
func testActionConfigWithNoRelease(log vzlog.VerrazzanoLogger, settings *cli.EnvSettings, namespace string) (*action.Configuration, error) {
	return vzhelm.CreateActionConfig(false, releaseName, release.StatusUninstalled, log, createRelease)
}
