// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package actionspi

import (
	"github.com/verrazzano/verrazzano-modules/common/pkg/vzlog"
	modulesv1alpha1 "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// HelmInfo contains all the information need to manage the lifecycle of Helm releases
type HelmInfo struct {
	// HelmRelease contains Helm release information
	*modulesv1alpha1.HelmRelease

	// CharDir is the local file system chart directory
	ChartDir string
}

// HandlerContext contains the handler contexts for the API handler methods
type HandlerContext struct {
	ctrlclient.Client
	Log    vzlog.VerrazzanoLogger
	DryRun bool
}

// ActionHandlers
type ActionHandlers struct {
	InstallActionHandler   LifecycleActionHandler
	UninstallActionHandler LifecycleActionHandler
	UpdateActionHandler    LifecycleActionHandler
	UpgradeActionHandler   LifecycleActionHandler
	ReconcileActionHandler LifecycleActionHandler
}

type HandlerConfig struct {
	HelmInfo
	CR     interface{}
	Scheme *runtime.Scheme
}

type LifecycleActionHandler interface {
	// GetActionName returns the action name
	GetActionName() string

	// Init initializes the component Hekn information
	Init(context HandlerContext, config HandlerConfig) (ctrl.Result, error)

	// IsActionNeeded returns true if action is needed
	IsActionNeeded(context HandlerContext) (bool, ctrl.Result, error)

	// PreAction does lifecycle pre-Action
	PreAction(context HandlerContext) (ctrl.Result, error)

	// PreActionUpdateStatus does the lifecycle pre-Action status update
	PreActionUpdateStatus(context HandlerContext) (ctrl.Result, error)

	// IsPreActionDone returns true if pre-Action done
	IsPreActionDone(context HandlerContext) (bool, ctrl.Result, error)

	// ActionUpdateStatus does the lifecycle action status update
	ActionUpdateStatus(context HandlerContext) (ctrl.Result, error)

	// DoAction does the lifecycle Action
	DoAction(context HandlerContext) (ctrl.Result, error)

	// IsActionDone returns true if action is done
	IsActionDone(context HandlerContext) (bool, ctrl.Result, error)

	// PostActionUpdateStatus does the lifecycle post-Action status update
	PostActionUpdateStatus(context HandlerContext) (ctrl.Result, error)

	// PostAction does lifecycle post-Action
	PostAction(context HandlerContext) (ctrl.Result, error)

	// IsPostActionDone returns true if action is done
	IsPostActionDone(context HandlerContext) (bool, ctrl.Result, error)

	// CompletedActionUpdateStatus does the lifecycle completed Action status update
	CompletedActionUpdateStatus(context HandlerContext) (ctrl.Result, error)
}
