// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package module

import (
	"fmt"
	actionspi "github.com/verrazzano/verrazzano-modules/common/actionspi"
	moduleplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/config"
	vzhelm "github.com/verrazzano/verrazzano/pkg/helm"
	"os"
	"path/filepath"
)

func loadHelmInfo(cr *moduleplatform.Module) (actionspi.HelmInfo, error) {
	chartDir := lookupChartDir(cr)
	isChartFound, err := isFileExist(chartDir)
	if err != nil {
		return actionspi.HelmInfo{}, err
	}
	if !isChartFound {
		return actionspi.HelmInfo{}, fmt.Errorf("FileNotFound at %s/Chart.yaml", chartDir)
	}
	chartInfo, err := vzhelm.GetChartInfo(chartDir)
	if err != nil {
		return actionspi.HelmInfo{}, err
	}

	helmInfo := actionspi.HelmInfo{
		HelmRelease: &moduleplatform.HelmRelease{
			Name:      cr.Name,
			Namespace: cr.Spec.TargetNamespace,
			ChartInfo: moduleplatform.HelmChart{
				Name:    chartInfo.Name,
				Version: chartInfo.Version,
				Path:    lookupChartDir(cr),
			},
			Overrides: nil,
		},
	}
	return helmInfo, nil
}

func lookupChartDir(mod *moduleplatform.Module) string {
	config := config.Get()
	chartpath := filepath.Join(config.ChartsDir, lookupChartLeafDirName(mod))
	return chartpath
}

func lookupChartLeafDirName(mod *moduleplatform.Module) string {
	var dir string

	switch mod.Spec.ModuleName {
	case string(moduleplatform.CalicoLifecycleClass):
		dir = "calico"
	case string(moduleplatform.CCMLifecycleClass):
		dir = "ccm"
	case string(moduleplatform.HelmLifecycleClass):
		dir = filepath.Join("vz-test", mod.Spec.Version)
		if mod.Spec.Version == "" {
			dir = filepath.Join("vz-test", "0.1.0")
		}
	}
	return dir
}

func isFileExist(chartDir string) (bool, error) {
	_, err := os.Stat(chartDir)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
