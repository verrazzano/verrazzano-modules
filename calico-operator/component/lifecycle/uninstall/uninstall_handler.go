// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package uninstall

import (
	compspi "github.com/verrazzano/verrazzano-modules/common/helm_component/spi"
	ctrl "sigs.k8s.io/controller-runtime"

	vzhelm "github.com/verrazzano/verrazzano/pkg/helm"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"

	"github.com/verrazzano/verrazzano/platform-operator/constants"
	helmcomp "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/helm"
)

type helmComponentAdapter struct {
	helmcomp.HelmComponent
	HelmInfo *compspi.HelmInfo
	chartDir string
}

var _ compspi.LifecycleActionHandler = &helmComponentAdapter{}

func NewComponent() compspi.LifecycleActionHandler {
	return &helmComponentAdapter{}
}

// Init initializes the component with Helm chart information
func (h *helmComponentAdapter) Init(_ spi.ComponentContext, HelmInfo *compspi.HelmInfo) (ctrl.Result, error) {
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
func (h helmComponentAdapter) PreAction(context spi.ComponentContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// IsPreActionDone returns true if pre-action done
func (h helmComponentAdapter) IsPreActionDone(context spi.ComponentContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// DoAction installs the component using Helm
func (h helmComponentAdapter) DoAction(context spi.ComponentContext) (ctrl.Result, error) {
	err := vzhelm.Uninstall(context.Log(), h.ReleaseName, h.ChartNamespace, context.IsDryRun())
	return ctrl.Result{}, err
}

// IsActionDone Indicates whether a component is installed and ready
func (h helmComponentAdapter) IsActionDone(context spi.ComponentContext) (bool, ctrl.Result, error) {
	if context.IsDryRun() {
		context.Log().Debugf("IsReady() dry run for %s", h.ReleaseName)
		return true, ctrl.Result{}, nil
	}

	deployed, err := vzhelm.IsReleaseDeployed(h.ReleaseName, h.ChartNamespace)
	if err != nil {
		context.Log().ErrorfThrottled("Error occurred checking release deloyment: %v", err.Error())
		return false, ctrl.Result{}, err
	}

	return !deployed, ctrl.Result{}, nil
}

// PostAction does installation pre-action
func (h helmComponentAdapter) PostAction(context spi.ComponentContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// IsPostActionDone returns true if post-action done
func (h helmComponentAdapter) IsPostActionDone(context spi.ComponentContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}
