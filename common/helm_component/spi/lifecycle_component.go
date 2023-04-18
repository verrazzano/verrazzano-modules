// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package spi

import (
	modulesv1alpha1 "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	vzspi "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	ctrl "sigs.k8s.io/controller-runtime"
)

type LifecycleComponent interface {
	vzspi.Component
	Init(context vzspi.ComponentContext, chartInfo *HelmInfo) error
}

type HelmInfo struct {
	*modulesv1alpha1.HelmRelease
}

type ActionLifecycle interface {
	PreWork(context vzspi.ComponentContext) (ctrl.Result, error)
	IsPreWorkDone(context vzspi.ComponentContext) (bool, ctrl.Result, error)
	DoWork(context vzspi.ComponentContext) (ctrl.Result, error)
	IsWorkDone(context vzspi.ComponentContext) (bool, ctrl.Result, error)
	PostWork(context vzspi.ComponentContext) (ctrl.Result, error)
	IsPostWorkDone(context vzspi.ComponentContext) (bool, ctrl.Result, error)
	IsActionDone(context vzspi.ComponentContext) (bool, ctrl.Result, error)
}
