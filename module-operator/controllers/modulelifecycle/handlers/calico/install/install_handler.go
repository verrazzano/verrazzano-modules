// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package install

import (
	"github.com/verrazzano/verrazzano-modules/common/handlerspi"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/modulelifecycle/handlers/helm/install"
	ctrl "sigs.k8s.io/controller-runtime"
)

type CalicoHandler struct {
	install.HelmHandler
}

var (
	_ handlerspi.StateMachineHandler = &CalicoHandler{}
)

func NewHandler() handlerspi.StateMachineHandler {
	return &CalicoHandler{}
}

// PreAction does installation pre-action
func (h CalicoHandler) PreAction(ctx handlerspi.HandlerContext) (ctrl.Result, error) {

	// TODO - Do Calico specific work here
	ctx.Log.Progress("Doing custom Calico pre-install logic")

	return h.HelmHandler.PreAction(ctx)
}

// IsPreActionDone returns true if pre-action done
func (h CalicoHandler) IsPreActionDone(ctx handlerspi.HandlerContext) (bool, ctrl.Result, error) {

	// TODO - Do Calico specific work here

	return h.HelmHandler.IsPreActionDone(ctx)
}
