// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package install

import (
	compspi "github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/action_spi"
	"github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/helm"
	"helm.sh/helm/v3/pkg/release"
	ctrl "sigs.k8s.io/controller-runtime"

	vzhelm "github.com/verrazzano/verrazzano/pkg/helm"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"

	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	moduleplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano/pkg/log/vzlog"
	"github.com/verrazzano/verrazzano/platform-operator/constants"
	helmcomp "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/helm"
)

type Component struct {
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

// GetStatusConditions returns the CR status conditions for various lifecycle stages
func (h *Component) GetStatusConditions() compspi.StatusConditions {
	return compspi.StatusConditions{
		NotNeeded: moduleplatform.CondAlreadyInstalled,
		PreAction: moduleplatform.CondPreInstall,
		DoAction:  moduleplatform.CondInstallStarted,
		Completed: moduleplatform.CondInstallComplete,
	}
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

// IsActionNeeded returns true if install is needed
func (h Component) IsActionNeeded(context spi.ComponentContext) (bool, ctrl.Result, error) {
	installed, err := vzhelm.IsReleaseInstalled(h.ReleaseName, h.chartDir)
	if err != nil {
		context.Log().ErrorfThrottled("Error checking if Helm release installed for %s/%s", h.chartDir, h.ReleaseName)
		return true, ctrl.Result{}, err
	}
	return !installed, ctrl.Result{}, err
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
	helmOverrides, err := helm.BuildHelmOverrides(context.Log(), context.Client(), helmRelease.Name, helmRelease.Namespace, helmRelease.Overrides)
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
		context.Log().ErrorfThrottled("Error occurred checking release deployment: %v", err.Error())
		return false, ctrl.Result{}, err
	}
	if !deployed {
		return false, util.NewRequeueWithShortDelay(), nil
	}
	// check if helm release at the correct version
	if !h.releaseVersionMatches(context.Log()) {
		return false, util.NewRequeueWithShortDelay(), nil
	}

	// TODO check if release is ready (check deployments)
	return true, ctrl.Result{}, err
}

// PostAction does installation pre-action
func (h Component) PostAction(context spi.ComponentContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// IsPostActionDone returns true if post-action done
func (h Component) IsPostActionDone(context spi.ComponentContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}
