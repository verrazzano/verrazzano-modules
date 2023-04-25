// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package update

import (
	compspi "github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/action_spi"
	"github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/helm"
	"helm.sh/helm/v3/pkg/release"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"

	moduleplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano/pkg/log/vzlog"
	"github.com/verrazzano/verrazzano/platform-operator/constants"
	helmcomp "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/helm"
)

type Component struct {
	helmcomp.HelmComponent
	HelmInfo     *compspi.HelmInfo
	chartDir     string
	mclNamespace string
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
func (h *Component) Init(_ spi.ComponentContext, HelmInfo *compspi.HelmInfo, mlcNamespace string) (ctrl.Result, error) {
	h.HelmComponent = helmcomp.HelmComponent{
		ReleaseName:             HelmInfo.HelmRelease.Name,
		ChartDir:                h.chartDir,
		ChartNamespace:          HelmInfo.HelmRelease.Namespace,
		IgnoreNamespaceOverride: true,
		ImagePullSecretKeyname:  constants.GlobalImagePullSecName,
	}

	h.mclNamespace = mlcNamespace
	h.HelmInfo = HelmInfo
	return ctrl.Result{}, nil
}

// IsActionNeeded returns true if install is needed
func (h Component) IsActionNeeded(context spi.ComponentContext) (bool, ctrl.Result, error) {
	// TODO - return false until update implemented
	return false, ctrl.Result{}, nil
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
	return ctrl.Result{}, nil
}

// IsActionDone Indicates whether a component is installed and ready
func (h Component) IsActionDone(context spi.ComponentContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// PostAction does installation pre-action
func (h Component) PostAction(context spi.ComponentContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// IsPostActionDone returns true if post-action done
func (h Component) IsPostActionDone(context spi.ComponentContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}
