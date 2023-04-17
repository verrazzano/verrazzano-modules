// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package spi

import (
	helmcomp "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/helm"
	vzspi "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
)

type LifecycleComponent interface {
	vzspi.Component
	Init(context vzspi.ComponentContext,chartInfo *ChartInfo) error
}

type ChartInfo struct {
	helmcomp.HelmComponent

	// ChartVersion is the version of the helm chart
	ChartVersion string

	// ChartDir is the helm chart directory
	ChartDir string

	// ChartNamespace is the namespace passed to the helm command
	ChartNamespace string

	// RepositoryURL The name or URL of the repository, e.g., http://myrepo/vz/stable
	RepositoryURL string

	// ReleaseName is the helm chart release name
	ReleaseName string

	// JSONName is the josn name of the verrazzano component in CRD
	JSONName string
}
