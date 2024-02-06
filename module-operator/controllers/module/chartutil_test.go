// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package module

import (
	"fmt"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/config"
)

// TestLoadHelmInfo tests the loadHelmInfo function
// GIVEN a Module
// WHEN the loadHelmInfo is called
// THEN ensure that the correct HelmInfo is returned
func TestLoadHelmInfo(t *testing.T) {
	asserts := assert.New(t)

	const (
		rootDir = "../../manifests/charts"
	)
	tests := []struct {
		name        string
		moduleName  string
		expectedDir string
		version     string
	}{
		{
			name:        "test-ccm",
			moduleName:  "oci-ccm",
			expectedDir: path.Join(rootDir, "modules/oci-ccm/1.27.0"),
		},
		{
			name:        "test-calico",
			moduleName:  "calico",
			expectedDir: path.Join(rootDir, "modules/calico/3.25.1"),
		},
		{
			name:        "test-multus",
			moduleName:  "multus",
			expectedDir: path.Join(rootDir, "modules/multus/4.0.2"),
		},
		{
			name:        "test-metallb",
			moduleName:  "metallb",
			expectedDir: path.Join(rootDir, "modules/metallb/0.13.10"),
		},
		{
			name:        "test-kubevirt",
			moduleName:  "kubevirt",
			expectedDir: path.Join(rootDir, "modules/kubevirt/0.59.0"),
		},
		{
			name:        "test-rook",
			moduleName:  "rook",
			expectedDir: path.Join(rootDir, "modules/rook/1.12.3"),
		},
		{
			name:        "test-vz-test",
			moduleName:  "helm",
			expectedDir: path.Join(rootDir, "vz-test/0.1.0"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			c := config.Get()
			defer func() { config.Set(c) }()
			c.ChartsDir = rootDir
			config.Set(c)

			mod := moduleapi.Module{
				Spec: moduleapi.ModuleSpec{
					ModuleName: test.moduleName,
				},
			}
			info, err := loadHelmInfo(&mod)
			asserts.NoError(err)
			asserts.NotNil(info)
			asserts.Equal(test.expectedDir, info.ChartInfo.Path)
		})
	}
}

// TestMissingChartFileError tests the loadHelmInfo function where the file is missing
// GIVEN a Module
// WHEN the loadHelmInfo is called with the wrong root dir
// THEN ensure that the correct error is returned
func TestMissingChartFileError(t *testing.T) {
	asserts := assert.New(t)
	c := config.Get()
	defer func() { config.Set(c) }()
	c.ChartsDir = "missingdir"
	config.Set(c)

	mod := moduleapi.Module{
		Spec: moduleapi.ModuleSpec{
			ModuleName: "oci-ccm",
		},
	}
	_, err := loadHelmInfo(&mod)
	asserts.Error(err)
	asserts.Contains(err.Error(), "NotFound")
}

// TestLookupChartLeafDirName tests the lookup of the chart leaf directory name
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
			name:        "test-ccm-1.25.0",
			moduleName:  "oci-ccm",
			version:     "1.25.0",
			expectedDir: "modules/oci-ccm/1.25.0",
		},
		{
			name:        "test-ccm",
			moduleName:  "oci-ccm",
			expectedDir: "modules/oci-ccm/1.27.0",
		},
		{
			name:        "test-calico-default",
			moduleName:  "calico",
			expectedDir: "modules/calico/3.27.0",
		},
		{
			name:        "test-calico-version",
			moduleName:  "calico",
			version:     "v3.25.0",
			expectedDir: "modules/calico/3.25.0",
		},
		{
			name:        "test-calico-version-no-v",
			moduleName:  "calico",
			version:     "3.25.0",
			expectedDir: "modules/calico/3.25.0",
		},
		{
			name:        "test-calico-version-v3.25.1",
			moduleName:  "calico",
			version:     "v3.25.1",
			expectedDir: "modules/calico/3.25.1",
		},
		{
			name:        "test-calico-version-3.25.1",
			moduleName:  "calico",
			version:     "3.25.1",
			expectedDir: "modules/calico/3.25.1",
		},
		{
			name:        "test-multus",
			moduleName:  "multus",
			expectedDir: "modules/multus/4.0.2",
		},
		{
			name:        "test-multus-version",
			moduleName:  "multus",
			version:     "4.0.1",
			expectedDir: "modules/multus/4.0.1",
		},
		{
			name:        "test-multus-version",
			moduleName:  "multus",
			version:     "4.0.2",
			expectedDir: "modules/multus/4.0.2",
		},
		{
			name:        "test-metallb",
			moduleName:  "metallb",
			expectedDir: "modules/metallb/0.13.10",
		},
		{
			name:        "test-metallb-version",
			moduleName:  "metallb",
			version:     "0.13.10",
			expectedDir: "modules/metallb/0.13.10",
		},
		{
			name:        "test-kubevirt",
			moduleName:  "kubevirt",
			expectedDir: "modules/kubevirt/0.59.0",
		},
		{
			name:        "test-kubevirt-version",
			moduleName:  "kubevirt",
			version:     "0.58.0",
			expectedDir: "modules/kubevirt/0.58.0",
		},
		{
			name:        "test-kubevirt-version",
			moduleName:  "kubevirt",
			version:     "0.59.0",
			expectedDir: "modules/kubevirt/0.59.0",
		},
		{
			name:        "test-rook",
			moduleName:  "rook",
			expectedDir: "modules/rook/1.12.3",
		},
		{
			name:        "test-rook-version",
			moduleName:  "rook",
			version:     "1.12.3",
			expectedDir: "modules/rook/1.12.3",
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

// TestLookupChartDir tests the lookup of the chart directory name
// GIVEN a Module
// WHEN the lookupChartLeafDirName is called
// THEN ensure that the correct directory name is returned
func TestLookupChartDir(t *testing.T) {
	asserts := assert.New(t)
	const (
		rootDir = "/root"
		calico  = "calico"
		version = "1.26.0"
	)

	c := config.Get()
	defer func() { config.Set(c) }()
	c.ChartsDir = rootDir
	config.Set(c)

	mod := moduleapi.Module{
		Spec: moduleapi.ModuleSpec{
			ModuleName: calico,
			Version:    version,
		},
	}
	dir := lookupChartDir(&mod)
	asserts.Equal(fmt.Sprintf("%s/modules/%s/%s", rootDir, calico, version), dir)
}
