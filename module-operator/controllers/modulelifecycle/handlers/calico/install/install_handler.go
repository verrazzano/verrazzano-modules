package install

import (
	"github.com/verrazzano/verrazzano-modules/common/actionspi"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/modulelifecycle/handlers/helm/install"
	ctrl "sigs.k8s.io/controller-runtime"
)

type CalicoHandler struct {
	install.HelmHandler
}

var (
	_ actionspi.LifecycleActionHandler = &CalicoHandler{}
)

func NewHandler() actionspi.LifecycleActionHandler {
	return &CalicoHandler{}
}

// PreAction does installation pre-action
func (h CalicoHandler) PreAction(ctx actionspi.HandlerContext) (ctrl.Result, error) {

	// TODO - Do Calico specific work here
	ctx.Log.Progress("Doing custom Calico pre-install logic")

	return h.HelmHandler.PreAction(ctx)
}

// IsPreActionDone returns true if pre-action done
func (h CalicoHandler) IsPreActionDone(ctx actionspi.HandlerContext) (bool, ctrl.Result, error) {

	// TODO - Do Calico specific work here

	return h.HelmHandler.IsPreActionDone(ctx)
}
