// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConfigDefaults tests the config default values
// GIVEN a new OperatorConfig object
//
//	WHEN I call New
//	THEN the value returned are correct defaults
func TestConfigDefaults(t *testing.T) {
	// Reset singleton
	instance = nil

	asserts := assert.New(t)
	conf := Get()
	asserts.Equal("/etc/webhook/certs", conf.CertDir, "CertDir is incorrect")
	asserts.True(conf.LeaderElectionEnabled, "LeaderElectionEnabled is incorrect")
	asserts.Equal(":8080", conf.MetricsAddr, "MetricsAddr is incorrect")
	asserts.Equal("default", conf.LeaderElectionNamespace, "LeaderElectionNamespace default is not correct")
}

// TestConfigLeaderElectionNamespace tests the config default values
// GIVEN a new OperatorConfig object with a leader election namespace env var set
//
//	WHEN I call New
//	THEN the value returned are correct LE namespace
func TestConfigLeaderElectionNamespace(t *testing.T) {
	asserts := assert.New(t)

	// Reset singleton
	instance = nil

	leNamespace := "verrazzano-install"
	os.Setenv(leaderElectionNamespaceVarName, leNamespace)
	defer func() {
		os.Unsetenv(leaderElectionNamespaceVarName)
	}()

	conf := Get()
	asserts.Equal("/etc/webhook/certs", conf.CertDir, "CertDir is incorrect")
	asserts.True(conf.LeaderElectionEnabled, "LeaderElectionEnabled is incorrect")
	asserts.Equal(":8080", conf.MetricsAddr, "MetricsAddr is incorrect")
	asserts.Equal(leNamespace, conf.LeaderElectionNamespace, "LeaderElectionNamespace default is not correct")
}

// TestSetConfig tests setting config values
// GIVEN an OperatorConfig object with non-default values
//
//		WHEN I call Set
//		THEN Get returns the correct values
//	    Able to override variables
func TestSetConfig(t *testing.T) {
	// Reset singleton
	instance = nil

	asserts := assert.New(t)
	Set(OperatorConfig{
		CertDir:                 "/test/certs",
		MetricsAddr:             "1111",
		LeaderElectionEnabled:   false,
		LeaderElectionNamespace: "myns",
	})
	conf := Get()
	asserts.Equal("/test/certs", conf.CertDir, "CertDir is incorrect")
	asserts.False(conf.LeaderElectionEnabled, "LeaderElectionEnabled is incorrect")
	asserts.Equal("1111", conf.MetricsAddr, "MetricsAddr is incorrect")
	asserts.Equal("myns", conf.LeaderElectionNamespace, "LeaderElectionNamespace is not correct")
}
