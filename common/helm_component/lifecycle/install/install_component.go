// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package install

import (
	"github.com/verrazzano/verrazzano-modules/common/helm_component/helm"
	compspi "github.com/verrazzano/verrazzano-modules/common/helm_component/spi"
	"helm.sh/helm/v3/pkg/release"

	vzhelm "github.com/verrazzano/verrazzano/pkg/helm"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"

	"github.com/verrazzano/verrazzano/pkg/log/vzlog"
	"github.com/verrazzano/verrazzano/platform-operator/constants"
	helmcomp "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/helm"
	"k8s.io/apimachinery/pkg/runtime"
)

type helmComponentAdapter struct {
	helmcomp.HelmComponent
	HelmInfo *compspi.HelmInfo
	chartDir string
}

// upgradeFuncSig is a function needed for unit test override
type upgradeFuncSig func(log vzlog.VerrazzanoLogger, releaseOpts *helm.HelmReleaseOpts, wait bool, dryRun bool) (*release.Release, error)

var (
	_ compspi.LifecycleComponent = &helmComponentAdapter{}

	upgradeFunc upgradeFuncSig = helm.UpgradeRelease
)

func NewComponent(chartDir string) compspi.LifecycleComponent {
	return &helmComponentAdapter{
		chartDir: chartDir,
	}
}

// Init initializes the component with Helm chart information
func (h *helmComponentAdapter) Init(_ spi.ComponentContext, HelmInfo *compspi.HelmInfo) error {
	h.HelmComponent = helmcomp.HelmComponent{
		ReleaseName:             HelmInfo.HelmRelease.Name,
		ChartDir:                h.chartDir,
		ChartNamespace:          HelmInfo.HelmRelease.Namespace,
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
	helmOverrides, err := helm.ConvertToHelmOverrides(context.Log(), context.Client(), helmRelease.Name, helmRelease.Namespace, helmRelease.Overrides)
	if err != nil {
		return err
	}
	var opts = &helm.HelmReleaseOpts{
		RepoURL:      helmRelease.Repository.URI,
		ReleaseName:  h.ReleaseName,
		Namespace:    h.ChartNamespace,
		ChartPath:    helmRelease.ChartInfo.Path,
		ChartVersion: helmRelease.ChartInfo.Version,
		Overrides:    helmOverrides,
		// TBD -- pull from a secret ref?
		//Username:     "",
		//Password:     "",
	}
	_, err = upgradeFunc(context.Log(), opts, h.WaitForInstall, context.IsDryRun())
	return err
}

// IsReady Indicates whether a component is available and ready
func (h helmComponentAdapter) IsReady(context spi.ComponentContext) bool {
	if context.IsDryRun() {
		context.Log().Debugf("IsReady() dry run for %s", h.ReleaseName)
		return true
	}

	deployed, err := vzhelm.IsReleaseDeployed(h.ReleaseName, h.ChartNamespace)
	if err != nil {
		context.Log().ErrorfThrottled("Error occurred checking release deloyment: %v", err.Error())
		return false
	}

	releaseMatches := h.releaseVersionMatches(context.Log())

	// The helm release exists and is at the correct version
	return deployed && releaseMatches
}

func (h helmComponentAdapter) Upgrade(context spi.ComponentContext) error {
	return h.Install(context)
}

func (h helmComponentAdapter) releaseVersionMatches(log vzlog.VerrazzanoLogger) bool {
	releaseChartVersion, err := helm.GetReleaseChartVersion(h.ReleaseName, h.ChartNamespace)
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
