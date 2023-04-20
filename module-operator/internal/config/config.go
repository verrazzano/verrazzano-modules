// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package config

import "os"

const leaderElectionNamespaceVarName = "LEADER_ELECTION_NAMESPACE"

// OperatorConfig specifies the Verrazzano Platform Operator Config
type OperatorConfig struct {

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
		instance = &OperatorConfig{
			CertDir:                 "/etc/webhook/certs",
			MetricsAddr:             ":8080",
			LeaderElectionEnabled:   true,
			LeaderElectionNamespace: GetWorkingNamespace(),
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
