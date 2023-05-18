// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package common

import (
	"context"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/handlerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type BaseHandler struct {
	Config                handlerspi.StateMachineHandlerConfig
	ModuleCR              *moduleapi.Module
	Action                moduleapi.ModuleActionType
	ModuleActionName      string
	ModuleActionNamespace string
}

func (h *BaseHandler) GetActionName() string {
	//TODO implement me
	return "unknown"
}

// Init initializes the handler with Helm chart information
func (h *BaseHandler) Init(_ handlerspi.HandlerContext, config handlerspi.StateMachineHandlerConfig, action moduleapi.ModuleActionType) (ctrl.Result, error) {
	h.Config = config
	h.ModuleCR = config.CR.(*moduleapi.Module)
	h.ModuleActionName = DeriveModuleLifeCycleName(h.ModuleCR.Name, moduleapi.HelmLifecycleClass, action)
	h.ModuleActionNamespace = h.ModuleCR.Namespace
	h.Action = action
	return ctrl.Result{}, nil
}

// IsWorkNeeded returns true if install is needed
func (h BaseHandler) IsWorkNeeded(ctx handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// PreWorkUpdateStatus does the lifecycle pre-Action status update
func (h *BaseHandler) PreWorkUpdateStatus(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// PreWork does the pre-work
func (h BaseHandler) PreWork(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// DoWorkUpdateStatus does the status update
func (h *BaseHandler) DoWorkUpdateStatus(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// DoWork does the main work for the Module lifecycle operator
func (h BaseHandler) DoWork(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	// Create ModuleAction
	moduleAction := moduleapi.ModuleAction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      h.ModuleActionName,
			Namespace: h.ModuleCR.Namespace,
		},
	}
	_, err := controllerutil.CreateOrUpdate(context.TODO(), ctx.Client, &moduleAction, func() error {
		err := h.mutateModuleAction(&moduleAction)
		if err != nil {
			return err
		}
		return controllerutil.SetControllerReference(h.ModuleCR, &moduleAction, h.Config.Scheme)
	})

	return ctrl.Result{}, err
}

func (h BaseHandler) mutateModuleAction(moduleAction *moduleapi.ModuleAction) error {
	moduleAction.Spec.ModuleClassName = moduleapi.ModuleClassType(h.ModuleCR.Spec.ModuleName)
	moduleAction.Spec.Action = h.Action
	moduleAction.Spec.Version = h.ModuleCR.Spec.Version
	moduleAction.Spec.Installer.HelmRelease = h.Config.HelmInfo.HelmRelease
	moduleAction.Spec.Installer.HelmRelease.Overrides = h.ModuleCR.Spec.Overrides
	return nil
}

// IsWorkDone returns true if the module work is done
func (h BaseHandler) IsWorkDone(ctx handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	if ctx.DryRun {
		return true, ctrl.Result{}, nil
	}

	moduleAction, err := h.GetModuleLifecycle(ctx)
	if err != nil {
		return false, util.NewRequeueWithShortDelay(), nil
	}
	if moduleAction.Status.State == moduleapi.StateCompleted || moduleAction.Status.State == moduleapi.StateNotNeeded {
		return true, ctrl.Result{}, nil
	}
	ctx.Log.Progressf("Waiting for ModuleAction %s to be completed", h.ModuleActionName)
	return false, ctrl.Result{}, nil
}

// PostWork does post-action
func (h BaseHandler) PostWork(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	if ctx.DryRun {
		return ctrl.Result{}, nil
	}

	if err := h.DeleteModuleLifecycle(ctx); err != nil {
		return util.NewRequeueWithShortDelay(), nil
	}
	return ctrl.Result{}, nil
}

// PostWorkUpdateStatus does installation post-action status update
func (h BaseHandler) PostWorkUpdateStatus(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// WorkCompletedUpdateStatus does the lifecycle completed Action status update
func (h *BaseHandler) WorkCompletedUpdateStatus(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// UpdateDoneStatus does the lifecycle status update with the action is done
func (h BaseHandler) UpdateDoneStatus(ctx handlerspi.HandlerContext, cond moduleapi.LifecycleCondition, state moduleapi.ModuleStateType, version string) (ctrl.Result, error) {
	AppendCondition(h.ModuleCR, string(cond), cond)
	h.ModuleCR.Status.State = state
	h.ModuleCR.Status.ObservedGeneration = h.ModuleCR.Generation
	if len(version) > 0 {
		h.ModuleCR.Status.Version = version
	}
	if err := ctx.Client.Status().Update(context.TODO(), h.ModuleCR); err != nil {
		return util.NewRequeueWithShortDelay(), nil
	}
	return ctrl.Result{}, nil
}

// UpdateStatus does the lifecycle pre-Action status update
func (h BaseHandler) UpdateStatus(ctx handlerspi.HandlerContext, cond moduleapi.LifecycleCondition, state moduleapi.ModuleStateType) (ctrl.Result, error) {
	AppendCondition(h.ModuleCR, string(cond), cond)
	h.ModuleCR.Status.State = state
	if err := ctx.Client.Status().Update(context.TODO(), h.ModuleCR); err != nil {
		return util.NewRequeueWithShortDelay(), nil
	}
	return ctrl.Result{}, nil
}
