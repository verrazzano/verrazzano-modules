// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package helm

import (
	"github.com/stretchr/testify/assert"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"testing"
)

const (
	releaseName = "test-release"
	namespace   = "test-ns"
	version     = "0.1.0"
)

// TestGetReleaseChartVersion tests getting the chart version from the Helm release
// GIVEN a Helm release
// WHEN I call TestGetReleaseChartVersion
// THEN the correct chart version is returned
func TestGetReleaseChartVersion(t *testing.T) {
	assert := assert.New(t)

	SetActionConfigFunction(fakeActionConfigWithRelease)
	defer SetDefaultActionConfigFunction()

	relVersion, err := GetReleaseChartVersion(releaseName, namespace)

	assert.NoError(err, "GetReleaseChartVersion returned an error")
	assert.Equal(version, relVersion)
}

// TestGetReleaseChartVersionNotFound tests a not found error when getting the chart version from a release
// GIVEN a Helm release that does not exist
// WHEN I call TestGetReleaseChartVersion
// THEN the correct not found error is returned
func TestGetReleaseChartVersionNotFound(t *testing.T) {
	assert := assert.New(t)

	SetActionConfigFunction(fakeActionConfigWithNoRelease)
	defer SetDefaultActionConfigFunction()

	_, err := GetReleaseChartVersion(releaseName, namespace)

	assert.NoError(err, "GetReleaseChartVersion returned an error")
	assert.Equal(ReleaseNotFound, err.Error(), "GetReleaseChartVersion should have returned not found")
}

// fakeActionConfigWithRelease is a fake action that returns an installed Helm release
func fakeActionConfigWithRelease(log vzlog.VerrazzanoLogger, settings *cli.EnvSettings, namespace string) (*action.Configuration, error) {
	return CreateActionConfig(true, releaseName, release.StatusDeployed, log, createRelease)
}

// fakeActionConfigWithNoRelease is a fake action that returns an uninstalled Helm release
func fakeActionConfigWithNoRelease(log vzlog.VerrazzanoLogger, settings *cli.EnvSettings, namespace string) (*action.Configuration, error) {
	return CreateActionConfig(false, releaseName, release.StatusUninstalled, log, createRelease)
}
