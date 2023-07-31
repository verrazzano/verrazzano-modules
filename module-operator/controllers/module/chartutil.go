// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package module

import (
	"fmt"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/handlerspi"
	"os"
	"path/filepath"
	"strings"

	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/config"
	"github.com/verrazzano/verrazzano-modules/pkg/helm"
)

func loadHelmInfo(cr *moduleapi.Module) (handlerspi.HelmInfo, error) {
	chartDir := lookupChartDir(cr)
	isChartFound, err := doesFileExist(chartDir)
	if err != nil {
		return handlerspi.HelmInfo{}, err
	}
	if !isChartFound {
		return handlerspi.HelmInfo{}, fmt.Errorf("FileNotFound at %s/Chart.yaml", chartDir)
	}
	chartInfo, err := helm.GetChartInfo(chartDir)
	if err != nil {
		return handlerspi.HelmInfo{}, err
	}

	helmInfo := handlerspi.HelmInfo{
		HelmRelease: &handlerspi.HelmRelease{
			Name:      cr.Name,
			Namespace: cr.Spec.TargetNamespace,
			ChartInfo: handlerspi.HelmChart{
				Name:    chartInfo.Name,
				Version: chartInfo.Version,
				Path:    lookupChartDir(cr),
			},
		},
	}
	return helmInfo, nil
}

func lookupChartDir(mod *moduleapi.Module) string {
	config := config.Get()
	chartpath := filepath.Join(config.ChartsDir, lookupChartLeafDirName(mod))
	return chartpath
}

func lookupChartLeafDirName(mod *moduleapi.Module) string {
	var dir string
	version := strings.TrimPrefix(mod.Spec.Version, "v")
	switch mod.Spec.ModuleName {
	case string(moduleapi.CalicoModuleClass):
		if version == "" {
			version = "3.25.0"
		}
		dir = filepath.Join("modules/calico", version)
	case string(moduleapi.CCMModuleClass):
		if version == "" {
			version = "1.25.0"
		}
		dir = filepath.Join("modules/oci-ccm", version)
	case string(moduleapi.MultusModuleClass):
		if version == "" {
			version = "4.0.1"
		}
		dir = filepath.Join("modules/multus", version)
	case string(moduleapi.HelmModuleClass):
		if version == "" {
			version = "0.1.0"
		}
		dir = filepath.Join("vz-test", version)
	default:
		// default to empty string
	}
	return dir
}

func doesFileExist(chartDir string) (bool, error) {
	_, err := os.Stat(chartDir)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
