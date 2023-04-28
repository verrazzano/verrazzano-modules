// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package module

import (
	compspi "github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/action_spi"
	moduleplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/config"
	"path/filepath"
)

func loadHelmInfo(cr *moduleplatform.Module) compspi.HelmInfo {
	//	chartName := lookupChartName(cr)

	helmInfo := compspi.HelmInfo{
		HelmRelease: &moduleplatform.HelmRelease{
			Name:      cr.Name,
			Namespace: cr.Spec.TargetNamespace,
			ChartInfo: moduleplatform.HelmChart{
				Name:    "vz-integration-operator",
				Version: "0.1.0",
				Path:    lookupChartPath(cr),
			},
			Overrides: nil,
		},
	}
	return helmInfo
}

func lookupChartPath(mod *moduleplatform.Module) string {
	chartpath := filepath.Join(config.Get().ChartsDir, lookupChartName(mod))
	return chartpath
}

func lookupChartName(mod *moduleplatform.Module) string {
	var chartName string

	switch mod.Spec.ModuleName {
	case string(moduleplatform.CalicoLifecycleClass):
		chartName = "verrazzano-calico-operator"
	case string(moduleplatform.CCMLifecycleClass):
		chartName = "verrazzano-ccm-operator"
	case string("vz-test"):
		chartName = "vz-test"
	}
	return chartName
}
