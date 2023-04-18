// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package spi

import (
	modulesv1alpha1 "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	vzspi "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	ctrl "sigs.k8s.io/controller-runtime"
)

type HelmInfo struct {
	*modulesv1alpha1.HelmRelease
}

type LifecycleComponent struct {
	InstallAction   LifecycleAction
	UninstallAction LifecycleAction
	UpdateAction    LifecycleAction
	UpgradeAction   LifecycleAction
}

type LifecycleAction interface {
	// Init initializes the component helm information
	Init(context vzspi.ComponentContext, chartInfo *HelmInfo) (ctrl.Result, error)

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
