// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package common

import (
	"context"
	compspi "github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/action_spi"
	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	moduleplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano/platform-operator/constants"
	helmcomp "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/helm"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type BaseHandler struct {
	helmcomp.HelmComponent
	HelmInfo     *compspi.HelmInfo
	ChartDir     string
	MlcName      string
	MlcNamespace string
	ModuleCR     *moduleplatform.Module
	Action       moduleplatform.ActionType
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
func (h *BaseHandler) Init(_ spi.ComponentContext, HelmInfo *compspi.HelmInfo, mlcNamespace string, cr interface{}, action moduleplatform.ActionType) (ctrl.Result, error) {
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
	h.Action = action
	return ctrl.Result{}, nil
}

// DoAction installs the component using Helm
func (h BaseHandler) DoAction(ctx spi.ComponentContext) (ctrl.Result, error) {
	// Create ModuleLifecycle
	mlc := moduleplatform.ModuleLifecycle{
		ObjectMeta: metav1.ObjectMeta{
			Name:      h.MlcName,
			Namespace: h.ModuleCR.Namespace,
		},
	}
	_, err := controllerutil.CreateOrUpdate(context.TODO(), ctx.Client(), &mlc, func() error {
		err := h.mutateMLC(&mlc)

		//	TODO - figure out how to get scheme, also does controller ref need to be set
		//	return controllerutil.SetControllerReference(h.moduleCR, &mlc, h.Scheme)
		return err
	})

	return ctrl.Result{}, err
}

func (h BaseHandler) mutateMLC(mlc *moduleplatform.ModuleLifecycle) error {
	mlc.Spec.LifecycleClassName = moduleplatform.HelmLifecycleClass
	mlc.Spec.Action = h.Action
	mlc.Spec.Installer.HelmRelease = h.HelmInfo.HelmRelease
	return nil
}

// IsActionDone returns true if the module action is done
func (h BaseHandler) IsActionDone(ctx spi.ComponentContext) (bool, ctrl.Result, error) {
	if ctx.IsDryRun() {
		ctx.Log().Debugf("IsReady() dry run for %s", h.ReleaseName)
		return true, ctrl.Result{}, nil
	}

	mlc, err := h.GetModuleLifecycle(ctx)
	if err != nil {
		return false, util.NewRequeueWithShortDelay(), nil
	}
	if mlc.Status.State == moduleplatform.StateCompleted || mlc.Status.State == moduleplatform.StateNotNeeded {
		return true, ctrl.Result{}, nil
	}
	ctx.Log().Progressf("Waiting for ModuleLifecycle %s to be completed", h.MlcName)
	return false, ctrl.Result{}, nil
}

// PostAction does post-action
func (h BaseHandler) PostAction(ctx spi.ComponentContext) (ctrl.Result, error) {
	if ctx.IsDryRun() {
		ctx.Log().Debugf("IsReady() dry run for %s", h.ReleaseName)
		return ctrl.Result{}, nil
	}

	if err := h.DeleteModuleLifecycle(ctx); err != nil {
		return util.NewRequeueWithShortDelay(), nil
	}
	return ctrl.Result{}, nil
}
