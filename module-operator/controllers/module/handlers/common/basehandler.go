// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package common

import (
	"context"
	compspi "github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/action_spi"
	moduleplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano/platform-operator/constants"
	helmcomp "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/helm"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

type BaseHandler struct {
	helmcomp.HelmComponent
	HelmInfo     *compspi.HelmInfo
	ChartDir     string
	MlcName      string
	MlcNamespace string
	ModuleCR     *moduleplatform.Module
}

// GetStatusConditions returns the CR status conditions for various lifecycle stages
func (h *BaseHandler) GetStatusConditions() compspi.StatusConditions {
	return compspi.StatusConditions{
		NotNeeded: moduleplatform.CondAlreadyInstalled,
		PreAction: moduleplatform.CondPreInstall,
		DoAction:  moduleplatform.CondInstallStarted,
		Completed: moduleplatform.CondInstallComplete,
	}
}

// Init initializes the handler with Helm chart information
func (h *BaseHandler) Init(_ spi.ComponentContext, HelmInfo *compspi.HelmInfo, mlcNamespace string, cr interface{}) (ctrl.Result, error) {
	h.HelmComponent = helmcomp.HelmComponent{
		ReleaseName:             HelmInfo.HelmRelease.Name,
		ChartDir:                h.ChartDir,
		ChartNamespace:          HelmInfo.HelmRelease.Namespace,
		IgnoreNamespaceOverride: true,
		ImagePullSecretKeyname:  constants.GlobalImagePullSecName,
	}

	h.ModuleCR = cr.(*moduleplatform.Module)
	h.MlcName = DeriveModuleLifeCycleName(h.ModuleCR.Name, moduleplatform.HelmLifecycleClass, moduleplatform.InstallAction)
	h.MlcNamespace = mlcNamespace
	h.HelmInfo = HelmInfo
	return ctrl.Result{}, nil
}

func (h BaseHandler) GetModuleLifecycle(ctx spi.ComponentContext) (*moduleplatform.ModuleLifecycle, error) {
	mlc := moduleplatform.ModuleLifecycle{}
	nsn := types.NamespacedName{
		Name:      DeriveModuleLifeCycleName(h.ModuleCR.Name, moduleplatform.HelmLifecycleClass, moduleplatform.InstallAction),
		Namespace: h.ModuleCR.Namespace,
	}

	if err := ctx.Client().Get(context.TODO(), nsn, &mlc); err != nil {
		ctx.Log().Progressf("Retrying get for ModuleLifecycle %v: %v", nsn, err)
		return nil, err
	}
	return &mlc, nil
}
