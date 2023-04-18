// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package install

import (
	"github.com/verrazzano/verrazzano-modules/common/controllers/lifecycle"
	"github.com/verrazzano/verrazzano-modules/common/helm_component/helm"
	compspi "github.com/verrazzano/verrazzano-modules/common/helm_component/spi"
	"helm.sh/helm/v3/pkg/release"
	ctrl "sigs.k8s.io/controller-runtime"

	vzhelm "github.com/verrazzano/verrazzano/pkg/helm"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"

	"github.com/verrazzano/verrazzano/pkg/log/vzlog"
	"github.com/verrazzano/verrazzano/platform-operator/constants"
	helmcomp "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/helm"
)

type calicoComponentAdapter struct {
	lifecycle.Component
	helmcomp.HelmComponent
	HelmInfo *compspi.HelmInfo
	chartDir string
}

// upgradeFuncSig is a function needed for unit test override
type upgradeFuncSig func(log vzlog.VerrazzanoLogger, releaseOpts *helm.HelmReleaseOpts, wait bool, dryRun bool) (*release.Release, error)

var (
	_ compspi.LifecycleActionHandler = &Component{}

	upgradeFunc upgradeFuncSig = helm.UpgradeRelease
)

func NewComponent() compspi.LifecycleActionHandler {
	return &Component{}
}

// Init initializes the component with Helm chart information
func (h *Component) Init(_ spi.ComponentContext, HelmInfo *compspi.HelmInfo) (ctrl.Result, error) {
	h.HelmComponent = helmcomp.HelmComponent{
		ReleaseName:             HelmInfo.HelmRelease.Name,
		ChartDir:                h.chartDir,
		ChartNamespace:          HelmInfo.HelmRelease.Namespace,
		IgnoreNamespaceOverride: true,
		ImagePullSecretKeyname:  constants.GlobalImagePullSecName,
	}

	h.HelmInfo = HelmInfo
	return ctrl.Result{}, nil
}

// PreAction does installation pre-action
func (h Component) PreAction(context spi.ComponentContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// IsPreActionDone returns true if pre-action done
func (h Component) IsPreActionDone(context spi.ComponentContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// DoAction installs the component using Helm
func (h Component) DoAction(context spi.ComponentContext) (ctrl.Result, error) {
	// Perform a Helm install using the helm upgrade --install command
	helmRelease := h.HelmInfo.HelmRelease
	helmOverrides, err := helm.ConvertToHelmOverrides(context.Log(), context.Client(), helmRelease.Name, helmRelease.Namespace, helmRelease.Overrides)
	if err != nil {
		return ctrl.Result{}, err
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
	return ctrl.Result{}, err
}

// IsActionDone Indicates whether a component is installed and ready
func (h Component) IsActionDone(context spi.ComponentContext) (bool, ctrl.Result, error) {
	if context.IsDryRun() {
		context.Log().Debugf("IsReady() dry run for %s", h.ReleaseName)
		return true, ctrl.Result{}, nil
	}

	deployed, err := vzhelm.IsReleaseDeployed(h.ReleaseName, h.ChartNamespace)
	if err != nil {
		context.Log().ErrorfThrottled("Error occurred checking release deloyment: %v", err.Error())
		return false, ctrl.Result{}, err
	}

	releaseMatches := h.releaseVersionMatches(context.Log())

	// The helm release exists and is at the correct version
	return deployed && releaseMatches, ctrl.Result{}, nil
}

// PostAction does installation pre-action
func (h Component) PostAction(context spi.ComponentContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// IsPostActionDone returns true if post-action done
func (h Component) IsPostActionDone(context spi.ComponentContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}
