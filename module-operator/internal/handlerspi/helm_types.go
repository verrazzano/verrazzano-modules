// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package handlerspi

// HelmInfo contains all the information need to manage the  of Helm releases
type HelmInfo struct {
	// HelmRelease contains Helm release information
	*HelmRelease
}

// HelmRelease contains the HelmRelease information
type HelmRelease struct {
	Name       string              `json:"name"`
	Namespace  string              `json:"namespace,omitempty"`
	ChartInfo  HelmChart           `json:"chart,omitempty"`
	Repository HelmChartRepository `json:"repo,omitempty"`
}

// HelmChartRepository contains the HelmRelease information
type HelmChartRepository struct {
	Name              string `json:"name"`
	URI               string `json:"uri"`
	CredentialsSecret string `json:"credentialsSecret,omitempty"`
}

// HelmChart contains the local HelmChart information
type HelmChart struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Path    string `json:"path,omitempty"`
}

// ChartVersion contains the helm chart version
type ChartVersion struct {
	Name              string `json:"name"`
	DefaultVersion    string `json:"defaultVersion,omitempty"`
	SupportedVersions string `json:"supportedVersions,omitempty"`
}
