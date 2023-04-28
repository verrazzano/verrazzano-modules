// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package module

import (
	compspi "github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/action_spi"
	moduleplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/config"
	vzhelm "github.com/verrazzano/verrazzano/pkg/helm"
	"path/filepath"
)

func loadHelmInfo(cr *moduleplatform.Module) (compspi.HelmInfo, error) {
	chartDir := lookupChartDir(cr)
	chartInfo, err := vzhelm.GetChartInfo(chartDir)
	if err != nil {
		return compspi.HelmInfo{}, err
	}

	helmInfo := compspi.HelmInfo{
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
	chartpath := filepath.Join(config.ChartsDir, lookupChartName(mod))
	return chartpath
}

func lookupChartName(mod *moduleplatform.Module) string {
	var chartName string

	switch mod.Spec.ModuleName {
	case string(moduleplatform.CalicoLifecycleClass):
		chartName = "verrazzano-calico-operator"
	case string(moduleplatform.CCMLifecycleClass):
		chartName = "verrazzano-ccm-operator"
	case string(moduleplatform.HelmLifecycleClass):
		chartName = "vz-test"
	}
	return chartName
}
