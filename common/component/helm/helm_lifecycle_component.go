// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package helm

import (
	compspi "github.com/verrazzano/verrazzano-modules/common/component/spi"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"

	"github.com/verrazzano/verrazzano/pkg/log/vzlog"
	"github.com/verrazzano/verrazzano/platform-operator/constants"
	helmcomp "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/helm"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/runtime"
)

type helmComponentAdapter struct {
	helmcomp.HelmComponent
	HelmInfo *compspi.HelmInfo
	chartDir string
}

// upgradeFuncSig is a function needed for unit test override
type upgradeFuncSig func(log vzlog.VerrazzanoLogger, releaseOpts *HelmReleaseOpts, wait bool, dryRun bool) (*release.Release, error)

var (
	_ compspi.LifecycleComponent = &helmComponentAdapter{}

	upgradeFunc upgradeFuncSig = UpgradeRelease
)

func NewComponent(chartDir string) compspi.LifecycleComponent {
	return &helmComponentAdapter{
		chartDir: chartDir,
	}
}

// Init initializes the component with Helm chart information
func (h *helmComponentAdapter) Init(_ spi.ComponentContext, HelmInfo *compspi.HelmInfo) error {
	h.HelmComponent = helmcomp.HelmComponent{
		ReleaseName:             HelmInfo.ReleaseName,
		ChartDir:                h.chartDir,
		ChartNamespace:          HelmInfo.ChartNamespace,
		IgnoreNamespaceOverride: true,
		ImagePullSecretKeyname:  constants.GlobalImagePullSecName,
	}

	h.HelmInfo = HelmInfo

	//	chartURL := fmt.Sprintf("%s/%s", installer.HelmRelease.Repository.URI, HelmInfo.Path)

	return nil
}

// Install installs the component using Helm
func (h helmComponentAdapter) Install(context spi.ComponentContext) error {
	// Perform a Helm install using the helm upgrade --install command
	helmRelease := h.HelmInfo.HelmRelease
	helmOverrides, err := ConvertToHelmOverrides(context.Log(), context.Client(), helmRelease.Name, helmRelease.Namespace, helmRelease.Overrides)
	if err != nil {
		return err
	}
	var opts = &HelmReleaseOpts{
		//		RepoURL:      h.RepositoryURL,
		ReleaseName:  h.ReleaseName,
		Namespace:    h.ChartNamespace,
		ChartPath:    helmRelease.ChartInfo.Name,
		ChartVersion: helmRelease.ChartInfo.Version,
		Overrides:    helmOverrides,
		// TBD -- pull from a secret ref?
		//Username:     "",
		//Password:     "",
	}
	_, err = upgradeFunc(context.Log(), opts, h.WaitForInstall, context.IsDryRun())
	return err
}

func (h helmComponentAdapter) Upgrade(context spi.ComponentContext) error {
	return h.Install(context)
}

func (h helmComponentAdapter) releaseVersionMatches(log vzlog.VerrazzanoLogger) bool {
	releaseChartVersion, err := GetReleaseChartVersion(h.ReleaseName, h.ChartNamespace)
	if err != nil {
		log.ErrorfThrottled("Error occurred getting release chart version: %v", err.Error())
		return false
	}
	return h.HelmInfo.ChartInfo.Version == releaseChartVersion
}

// IsEnabled ModuleLifecycle objects are always enabled; if a Module is disabled the ModuleLifecycle resource doesn't exist
func (h helmComponentAdapter) IsEnabled(_ runtime.Object) bool {
	return true
}
