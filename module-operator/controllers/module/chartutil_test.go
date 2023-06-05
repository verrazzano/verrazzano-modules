// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package module

import (
	"testing"

	"github.com/stretchr/testify/assert"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
)

// TestLookupChartLeafDirName tests the lookup of the chart directory name
// GIVEN a Module
// WHEN the lookupChartLeafDirName is called
// THEN ensure that the correct directory name is returned
func TestLookupChartLeafDirName(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name        string
		moduleName  string
		expectedDir string
		version     string
	}{
		{
			name:        "test-ccm",
			moduleName:  "oci-ccm",
			expectedDir: "modules/oci-ccm/1.25.0",
		},
		{
			name:        "test-calico",
			moduleName:  "calico",
			expectedDir: "modules/calico/3.25.0",
		},
		{
			name:        "test-calico-version",
			moduleName:  "calico",
			version:     "v1.26.0",
			expectedDir: "modules/calico/1.26.0",
		},
		{
			name:        "test-calico-version-no-v",
			moduleName:  "calico",
			version:     "1.26.0",
			expectedDir: "modules/calico/1.26.0",
		},
		{
			name:        "test-vz-test",
			moduleName:  "helm",
			expectedDir: "vz-test/0.1.0",
		},
		{
			name:        "test-vz-test-version",
			moduleName:  "helm",
			version:     "0.1.1",
			expectedDir: "vz-test/0.1.1",
		},
		{
			name:        "test-unknown",
			moduleName:  "unknown",
			expectedDir: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mod := moduleapi.Module{
				Spec: moduleapi.ModuleSpec{
					ModuleName: test.moduleName,
					Version:    test.version,
				},
			}
			dir := lookupChartLeafDirName(&mod)
			asserts.Equal(test.expectedDir, dir)
		})
	}
}
