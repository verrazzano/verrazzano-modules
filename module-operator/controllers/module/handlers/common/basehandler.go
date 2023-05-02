// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package common

import (
	"context"
	actionspi "github.com/verrazzano/verrazzano-modules/common/actionspi"
	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	moduleplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type BaseHandler struct {
	Config       actionspi.HandlerConfig
	ModuleCR     *moduleplatform.Module
	Action       moduleplatform.ActionType
	MlcName      string
	MlcNamespace string
}

// Init initializes the handler with Helm chart information
func (h *BaseHandler) Init(_ spi.ComponentContext, config actionspi.HandlerConfig, action moduleplatform.ActionType) (ctrl.Result, error) {
	h.Config = config
	h.ModuleCR = config.CR.(*moduleplatform.Module)
	h.MlcName = DeriveModuleLifeCycleName(h.ModuleCR.Name, moduleplatform.HelmLifecycleClass, moduleplatform.InstallAction)
	h.MlcNamespace = h.ModuleCR.Namespace
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
		if err != nil {
			return err
		}
		return controllerutil.SetControllerReference(h.ModuleCR, &mlc, h.Config.Scheme)
	})

	return ctrl.Result{}, err
}

func (h BaseHandler) mutateMLC(mlc *moduleplatform.ModuleLifecycle) error {
	mlc.Spec.LifecycleClassName = moduleplatform.HelmLifecycleClass
	mlc.Spec.Action = h.Action
	mlc.Spec.Installer.HelmRelease = h.Config.HelmInfo.HelmRelease
	mlc.Spec.Installer.HelmRelease.Overrides = h.ModuleCR.Spec.Overrides
	return nil
}

// IsActionDone returns true if the module action is done
func (h BaseHandler) IsActionDone(ctx spi.ComponentContext) (bool, ctrl.Result, error) {
	if ctx.IsDryRun() {
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
		return ctrl.Result{}, nil
	}

	if err := h.DeleteModuleLifecycle(ctx); err != nil {
		return util.NewRequeueWithShortDelay(), nil
	}
	return ctrl.Result{}, nil
}

// UpdateStatus does the lifecycle pre-Action status update
func (h BaseHandler) UpdateStatusWithVersion(ctx spi.ComponentContext, cond moduleplatform.LifecycleCondition, state moduleplatform.ModuleStateType, version string) (ctrl.Result, error) {
	AppendCondition(h.ModuleCR, string(cond), cond)
	h.ModuleCR.Status.State = state
	if len(version) > 0 {
		h.ModuleCR.Status.Version = version
	}
	if err := ctx.Client().Status().Update(context.TODO(), h.ModuleCR); err != nil {
		return util.NewRequeueWithShortDelay(), nil
	}
	return ctrl.Result{}, nil
}

// UpdateStatus does the lifecycle pre-Action status update
func (h BaseHandler) UpdateStatus(ctx spi.ComponentContext, cond moduleplatform.LifecycleCondition, state moduleplatform.ModuleStateType) (ctrl.Result, error) {
	AppendCondition(h.ModuleCR, string(cond), cond)
	h.ModuleCR.Status.State = state
	if err := ctx.Client().Status().Update(context.TODO(), h.ModuleCR); err != nil {
		return util.NewRequeueWithShortDelay(), nil
	}
	return ctrl.Result{}, nil
}
