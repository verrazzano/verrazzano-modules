// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package config

import (
	"os"
	"path/filepath"
)

const (
	leaderElectionNamespaceVarName = "LEADER_ELECTION_NAMESPACE"
	rootDirEnv                     = "VZ_ROOT_DIR"
	defaultRootDir                 = "/home/verrazzano"
	manifestRelDir                 = "manifests"
	chartsRelDir                   = "manifests/charts/modules"
)

// OperatorConfig specifies the module operator config
type OperatorConfig struct {
	// The RootDir is the root directory on the image (or the local system when developing)
	RootDir string

	// The Manifest dir is the absolute directory path of the manifest
	ManifestDir string

	// The Charts dir is the absolute directory path of the charts
	ChartsDir string

	// The CertDir directory containing tls.crt and tls.key
	CertDir string

	// MetricsAddr is the address the metric endpoint binds to
	MetricsAddr string

	// LeaderElectionEnabled  enables/disables ensuring that there is only one active controller manager
	LeaderElectionEnabled bool

	// LeaderElectionNamespace the namespace to use for leader election
	LeaderElectionNamespace string
}

// The singleton instance of the operator config
var instance *OperatorConfig

// Set saves the operator config.  This should only be called at operator startup and during unit tests
func Set(config OperatorConfig) {
	instance = &OperatorConfig{}
	*instance = config
}

// Get returns the singleton instance of the operator config
func Get() OperatorConfig {
	if instance == nil {
		rootDir := os.Getenv(rootDirEnv)
		if len(rootDir) == 0 {
			rootDir = defaultRootDir
		}
		manifestDir := filepath.Join(rootDir, manifestRelDir)
		chartsDir := filepath.Join(rootDir, chartsRelDir)

		instance = &OperatorConfig{
			CertDir:                 "/etc/webhook/certs",
			MetricsAddr:             ":8080",
			LeaderElectionEnabled:   false,
			LeaderElectionNamespace: GetWorkingNamespace(),
			RootDir:                 rootDir,
			ManifestDir:             manifestDir,
			ChartsDir:               chartsDir,
		}
	}
	return *instance
}

func GetWorkingNamespace() string {
	workingNamespace, found := os.LookupEnv(leaderElectionNamespaceVarName)
	if !found {
		return "default"
	}
	return workingNamespace
}
