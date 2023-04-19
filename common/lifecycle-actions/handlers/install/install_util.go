// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package install

import (
	"github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/helm"
	"github.com/verrazzano/verrazzano/pkg/log/vzlog"
)

func (h Component) releaseVersionMatches(log vzlog.VerrazzanoLogger) bool {
	releaseChartVersion, err := helm.GetReleaseChartVersion(h.ReleaseName, h.ChartNamespace)
	if err != nil {
		log.ErrorfThrottled("Error occurred getting release chart version: %v", err.Error())
		return false
	}
	return h.HelmInfo.ChartInfo.Version == releaseChartVersion
}

// downloadChart will perform yum calls with specified arguments for operations
// The verbose field is just used for visibility of command in logging
func downloadChart(log vzlog.VerrazzanoLogger, verbose bool) (stdout []byte, stderr []byte, err error) {
	cmdArgs := []string{"install", "-y", "https://yum.oracle.com/repo/OracleLinux/OL8/olcne16/x86_64/getPackage/olcne-prometheus-chart-1.6.0-4.el8.x86_64.rpm"}
	cmd := exec.Command("yum", cmdArgs...)
	if verbose {
		log.Progressf("Running yum command: %s", cmd.String())
	}
	stdout, stderr, err = runner.Run(cmd)
	if err != nil {
		if verbose {
			log.Progressf("Failed running yum command %s: %s", cmd.String(), stderr)
		}
		return stdout, stderr, err
	}
	log.Debugf("yum %s succeeded: %s", operationName, stdout)
	return stdout, stderr, nil
}
