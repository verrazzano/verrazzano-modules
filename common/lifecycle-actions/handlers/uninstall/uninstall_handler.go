// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package uninstall

import (
	compspi "github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/action_spi"
	moduleplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	vzhelm "github.com/verrazzano/verrazzano/pkg/helm"
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
	return "uninstall"
}

// GetStatusConditions returns the CR status conditions for various lifecycle stages
func (h *Component) GetStatusConditions() compspi.StatusConditions {
	return compspi.StatusConditions{
		NotNeeded: moduleplatform.CondAlreadyUninstalled,
		PreAction: moduleplatform.CondPreUninstall,
		DoAction:  moduleplatform.CondUninstallStarted,
		Completed: moduleplatform.CondUninstallComplete,
	}
}

// IsActionNeeded returns true if uninstall is needed
func (h Component) IsActionNeeded(context spi.ComponentContext) (bool, ctrl.Result, error) {
	installed, err := vzhelm.IsReleaseInstalled(h.ReleaseName, h.Config.Namespace)
	if err != nil {
		context.Log().ErrorfThrottled("Error checking if Helm release installed for %s/%s", h.Config.ChartDir, h.ReleaseName)
		return true, ctrl.Result{}, err
	}
	return installed, ctrl.Result{}, err
}

// PreAction does installation pre-action
func (h Component) PreAction(context spi.ComponentContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// IsPreActionDone returns true if pre-action done
func (h Component) IsPreActionDone(context spi.ComponentContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// DoAction uninstalls the component using Helm
func (h Component) DoAction(context spi.ComponentContext) (ctrl.Result, error) {
	installed, err := vzhelm.IsReleaseInstalled(h.ReleaseName, h.Config.Namespace)
	if err != nil {
		context.Log().ErrorfThrottled("Error checking if Helm release installed for %s/%s", h.Config.ChartDir, h.ReleaseName)
		return ctrl.Result{}, err
	}
	if !installed {
		return ctrl.Result{}, err
	}

	err = vzhelm.Uninstall(context.Log(), h.ReleaseName, h.ChartNamespace, context.IsDryRun())
	return ctrl.Result{}, err
}

// IsActionDone Indicates whether a component is uninstalled
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

	return !deployed, ctrl.Result{}, nil
}

// PostAction does uninstall post-action
func (h Component) PostAction(context spi.ComponentContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// IsPostActionDone returns true if post-action done
func (h Component) IsPostActionDone(context spi.ComponentContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}
