// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package action_spi

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

// StatusConditions are conditions written to the CR Status during statemachine execution
type StatusConditions struct {
	NotNeeded modulesv1alpha1.LifecycleCondition
	PreAction modulesv1alpha1.LifecycleCondition
	DoAction  modulesv1alpha1.LifecycleCondition
	Completed modulesv1alpha1.LifecycleCondition
}

// ActionHandlers
type ActionHandlers struct {
	InstallAction   LifecycleActionHandler
	UninstallAction LifecycleActionHandler
	UpdateAction    LifecycleActionHandler
	UpgradeAction   LifecycleActionHandler
}

type HandlerConfig struct {
	HelmInfo
	CR     interface{}
	Scheme *runtime.Scheme
}

type LifecycleActionHandler interface {
	// GetActionName returns the action name
	GetActionName() string

	// GetStatusConditions returns the CR status conditions for various lifecycle stages
	GetStatusConditions() StatusConditions

	// Init initializes the component Hekn information
	Init(context vzspi.ComponentContext, config HandlerConfig) (ctrl.Result, error)

	// IsActionNeeded returns true if action is needed
	IsActionNeeded(context vzspi.ComponentContext) (bool, ctrl.Result, error)

	// PreAction does lifecycle pre-Action
	PreAction(context vzspi.ComponentContext) (ctrl.Result, error)

	// IsPreActionDone returns true if pre-Action done
	IsPreActionDone(context vzspi.ComponentContext) (bool, ctrl.Result, error)

	// DoAction does the lifecycle Action
	DoAction(context vzspi.ComponentContext) (ctrl.Result, error)

	// IsActionDone returns true if action is done
	IsActionDone(context vzspi.ComponentContext) (bool, ctrl.Result, error)

	// PostAction does lifecycle post-Action
	PostAction(context vzspi.ComponentContext) (ctrl.Result, error)

	// IsPostActionDone returns true if action is done
	IsPostActionDone(context vzspi.ComponentContext) (bool, ctrl.Result, error)
}
