// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package update

import (
	compspi "github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/action_spi"
	moduleplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano/platform-operator/constants"
	helmcomp "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/helm"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Component struct {
	helmcomp.HelmComponent
	Config compspi.HandlerConfig
	CR     *moduleplatform.ModuleLifecycle
}

var (
	_ compspi.LifecycleActionHandler = &Component{}
)

func NewComponent() compspi.LifecycleActionHandler {
	return &Component{}
}

// Init initializes the component with Helm chart information
func (h *Component) Init(_ spi.ComponentContext, config compspi.HandlerConfig) (ctrl.Result, error) {
	h.HelmComponent = helmcomp.HelmComponent{
		ReleaseName:             config.HelmInfo.HelmRelease.Name,
		ChartNamespace:          config.HelmInfo.HelmRelease.Namespace,
		ChartDir:                config.ChartDir,
		IgnoreNamespaceOverride: true,
		ImagePullSecretKeyname:  constants.GlobalImagePullSecName,
	}
	h.CR = config.CR.(*moduleplatform.ModuleLifecycle)
	h.Config = config
	return ctrl.Result{}, nil
}

// GetActionName returns the action name
func (h Component) GetActionName() string {
	return "update"
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
