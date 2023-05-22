// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package install

import (
	"github.com/verrazzano/verrazzano-modules/pkg/helm"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
)

func (h HelmHandler) releaseVersionMatches(log vzlog.VerrazzanoLogger) bool {
	releaseChartVersion, err := helm.GetReleaseChartVersion(h.HelmRelease.Name, h.HelmRelease.Namespace)
	if err != nil {
		log.ErrorfThrottled("Error occurred getting release chart version: %v", err.Error())
		return false
	}
	return h.BaseHandler.Config.HelmInfo.ChartInfo.Version == releaseChartVersion
}
