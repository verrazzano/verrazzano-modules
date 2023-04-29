// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package install

import (
	"github.com/verrazzano/verrazzano-modules/common/pkg/helm"
	"github.com/verrazzano/verrazzano/pkg/log/vzlog"
)

func (h Component) releaseVersionMatches(log vzlog.VerrazzanoLogger) bool {
	releaseChartVersion, err := helm.GetReleaseChartVersion(h.ReleaseName, h.ChartNamespace)
	if err != nil {
		log.ErrorfThrottled("Error occurred getting release chart version: %v", err.Error())
		return false
	}
	return h.Config.HelmInfo.ChartInfo.Version == releaseChartVersion
}
