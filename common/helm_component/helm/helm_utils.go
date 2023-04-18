// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package helm

import (
	"github.com/verrazzano/verrazzano/pkg/helm"
	"github.com/verrazzano/verrazzano/pkg/log/vzlog"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
)

func UpgradeRelease(log vzlog.VerrazzanoLogger, releaseOpts *HelmReleaseOpts, wait bool, dryRun bool) (*release.Release, error) {
	log.Infof("Upgrading release %s in namespace %s, chart %s, version %s, repoURL %s", releaseOpts.ReleaseName,
		releaseOpts.Namespace, releaseOpts.ChartPath, releaseOpts.ChartVersion, releaseOpts.RepoURL)
	settings := cli.New()
	settings.SetNamespace(releaseOpts.Namespace)

	chartOptions := action.ChartPathOptions{
		RepoURL:  releaseOpts.RepoURL,
		Version:  releaseOpts.ChartVersion,
		Password: releaseOpts.Username,
		Username: releaseOpts.Password,
	}
	chartPath := releaseOpts.ChartPath
	if chartPath == "" {
		var err error
		chartPath, err = chartOptions.LocateChart(releaseOpts.ChartPath, settings)
		if err != nil {
			return nil, err
		}
	}

	return helm.Upgrade(log, releaseOpts.ReleaseName, releaseOpts.Namespace, chartPath, wait, dryRun, releaseOpts.Overrides)
}
