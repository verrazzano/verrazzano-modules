// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package install

import (
	"context"
	compspi "github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/action_spi"
	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	"github.com/verrazzano/verrazzano-modules/common/pkg/helm"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/handlers/common"
	"helm.sh/helm/v3/pkg/release"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"

	moduleplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano/pkg/log/vzlog"
)

type Handler struct {
	BaseHandler common.BaseHandler
}

// upgradeFuncSig is a function needed for unit test override
type upgradeFuncSig func(log vzlog.VerrazzanoLogger, releaseOpts *helm.HelmReleaseOpts, wait bool, dryRun bool) (*release.Release, error)

var (
	_ compspi.LifecycleActionHandler = &Handler{}

	upgradeFunc upgradeFuncSig = helm.UpgradeRelease
)

func NewHandler() compspi.LifecycleActionHandler {
	return &Handler{}
}

// GetStatusConditions returns the CR status conditions for various lifecycle stages
func (h *Handler) GetStatusConditions() compspi.StatusConditions {
	return h.GetStatusConditions()
}

// Init initializes the component with Helm chart information
func (h *Handler) Init(ctx spi.ComponentContext, HelmInfo *compspi.HelmInfo, mlcNamespace string, cr interface{}) (ctrl.Result, error) {
	return h.Init(ctx, HelmInfo, mlcNamespace, cr)
}

// IsActionNeeded returns true if install is needed
func (h Handler) IsActionNeeded(ctx spi.ComponentContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil

	//installed, err := vzhelm.IsReleaseInstalled(h.ReleaseName, h.chartDir)
	//if err != nil {
	//	ctx.Log().ErrorfThrottled("Error checking if Helm release installed for %s/%s", h.chartDir, h.ReleaseName)
	//	return true, ctrl.Result{}, err
	//}
	//return !installed, ctrl.Result{}, err
}

// PreAction does installation pre-action
func (h Handler) PreAction(ctx spi.ComponentContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// IsPreActionDone returns true if pre-action done
func (h Handler) IsPreActionDone(ctx spi.ComponentContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

// DoAction installs the component using Helm
func (h Handler) DoAction(ctx spi.ComponentContext) (ctrl.Result, error) {
	// Create ModuleLifecycle
	mlc := moduleplatform.ModuleLifecycle{
		ObjectMeta: metav1.ObjectMeta{
			Name:      h.BaseHandler.MlcName,
			Namespace: h.BaseHandler.ModuleCR.Namespace,
		},
	}
	_, err := controllerutil.CreateOrUpdate(context.TODO(), ctx.Client(), &mlc, func() error {
		err := h.mutateMLC(&mlc)

		//	TODO - figure out how to get scheme, also does controller ref need to be set
		//	return controllerutil.SetControllerReference(h.moduleCR, &mlc, h.Scheme)
		return err
	})

	return ctrl.Result{}, err
}

func (h Handler) mutateMLC(mlc *moduleplatform.ModuleLifecycle) error {
	mlc.Spec.LifecycleClassName = moduleplatform.HelmLifecycleClass
	mlc.Spec.Action = moduleplatform.InstallAction
	mlc.Spec.Installer.HelmRelease = h.BaseHandler.HelmInfo.HelmRelease
	return nil
}

// IsActionDone Indicates whether a component is installed and ready
func (h Handler) IsActionDone(ctx spi.ComponentContext) (bool, ctrl.Result, error) {
	if ctx.IsDryRun() {
		ctx.Log().Debugf("IsReady() dry run for %s", h.BaseHandler.ReleaseName)
		return true, ctrl.Result{}, nil
	}

	mlc, err := h.BaseHandler.GetModuleLifecycle(ctx)
	if err != nil {
		return false, util.NewRequeueWithShortDelay(), nil
	}
	if mlc.Status.State == moduleplatform.StateCompleted || mlc.Status.State == moduleplatform.StateNotNeeded {
		return true, ctrl.Result{}, nil
	}
	ctx.Log().Progressf("Waiting for ModuleLifecycle %s to be completed", h.BaseHandler.MlcName)
	return false, ctrl.Result{}, nil
}

// PostAction does installation post-action
func (h Handler) PostAction(ctx spi.ComponentContext) (ctrl.Result, error) {
	if ctx.IsDryRun() {
		ctx.Log().Debugf("IsReady() dry run for %s", h.BaseHandler.ReleaseName)
		return ctrl.Result{}, nil
	}

	if err := h.BaseHandler.DeleteModuleLifecycle(ctx); err != nil {
		return util.NewRequeueWithShortDelay(), nil
	}
	return ctrl.Result{}, nil
}

// IsPostActionDone returns true if post-action done
func (h Handler) IsPostActionDone(ctx spi.ComponentContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}
