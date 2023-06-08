// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package install

import (
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/handlers/helm/install"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/handlerspi"
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

// PreWork does installation pre-work
func (h CalicoHandler) PreWork(ctx handlerspi.HandlerContext) result.Result {

	// TODO - Do Calico specific work here
	ctx.Log.Progress("Doing custom Calico pre-install logic")

	return h.HelmHandler.PreWork(ctx)
}
