// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package actionspi

import (
	modulesv1alpha1 "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	vzspi "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

// HelmInfo contains all the information need to manage the lifecycle of Helm releases
type HelmInfo struct {
	// HelmRelease contains Helm release information
	*modulesv1alpha1.HelmRelease

	// CharDir is the local file system chart directory
	ChartDir string
}

// ActionHandlers
type ActionHandlers struct {
	InstallActionHandler   LifecycleActionHandler
	UninstallActionHandler LifecycleActionHandler
	UpdateActionHandler    LifecycleActionHandler
	UpgradeActionHandler   LifecycleActionHandler
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
	Init(context vzspi.ComponentContext, config HandlerConfig) (ctrl.Result, error)

	// IsActionNeeded returns true if action is needed
	IsActionNeeded(context vzspi.ComponentContext) (bool, ctrl.Result, error)

	// PreAction does lifecycle pre-Action
	PreAction(context vzspi.ComponentContext) (ctrl.Result, error)

	// PreActionUpdateStatus does the lifecycle pre-Action status update
	PreActionUpdateStatus(context vzspi.ComponentContext) (ctrl.Result, error)

	// IsPreActionDone returns true if pre-Action done
	IsPreActionDone(context vzspi.ComponentContext) (bool, ctrl.Result, error)

	// ActionUpdateStatus does the lifecycle action status update
	ActionUpdateStatus(context vzspi.ComponentContext) (ctrl.Result, error)

	// DoAction does the lifecycle Action
	DoAction(context vzspi.ComponentContext) (ctrl.Result, error)

	// IsActionDone returns true if action is done
	IsActionDone(context vzspi.ComponentContext) (bool, ctrl.Result, error)

	// PostActionUpdateStatus does the lifecycle post-Action status update
	PostActionUpdateStatus(context vzspi.ComponentContext) (ctrl.Result, error)

	// PostAction does lifecycle post-Action
	PostAction(context vzspi.ComponentContext) (ctrl.Result, error)

	// IsPostActionDone returns true if action is done
	IsPostActionDone(context vzspi.ComponentContext) (bool, ctrl.Result, error)

	// CompletedActionUpdateStatus does the lifecycle completed Action status update
	CompletedActionUpdateStatus(context vzspi.ComponentContext) (ctrl.Result, error)
}
