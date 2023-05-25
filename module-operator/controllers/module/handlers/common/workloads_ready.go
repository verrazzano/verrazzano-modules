// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package common

import (
	"github.com/verrazzano/verrazzano-modules/pkg/helm"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
)

func CheckWorkLoadsReady(log vzlog.VerrazzanoLogger, releaseName string, namespace string) (bool, error) {
	// Get all the deployments, statefulsets, and daemonsets for this Helm release

	rel, err := helm.GetRelease(log, releaseName, namespace)
	if err != nil {
		return false, err
	}
	if rel.Manifest == "" {
		return false, err
	}

	return true, err
}
