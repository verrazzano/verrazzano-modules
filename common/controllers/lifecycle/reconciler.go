// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package lifecycle

import (
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/spi"
	"github.com/verrazzano/verrazzano-modules/common/k8s"
	vzplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"

	modulesv1alpha1 "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/common"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/modlifecycle/delegates"
	vzctrl "github.com/verrazzano/verrazzano-modules/module-operator/pkg/controller"
	"github.com/verrazzano/verrazzano/pkg/log/vzlog"
	vzspi "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
)

// Reconcile updates the Certificate
func (r Reconciler) Reconcile(spictx spi.ReconcileContext, u *unstructured.Unstructured) (ctrl.Result, error) {
	mlc := &vzplatform.ModuleLifecycle{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, mlc); err != nil {
		return ctrl.Result{}, err
	}
	ctx, err := vzspi.NewMinimalContext(r.Client, spictx.Log)
	if err != nil {
		return newRequeueWithDelay(), err
	}

	nsn := k8s.GetNamespacedName(mlc.ObjectMeta)
	if mlc.Generation == mlc.Status.ObservedGeneration {
		spictx.Log.Debugf("Skipping reconcile for %v, observed generation has not change", nsn)
		return newRequeueWithDelay(), err
	}
	if err := r.doReconcile(ctx, mlc); err != nil {
		return newRequeueWithDelay(), err
	}
	return ctrl.Result{}, nil
}

func newRequeueWithDelay() ctrl.Result {
	return vzctrl.NewRequeueWithDelay(3, 10, time.Second)
}

func (r *Reconciler) doReconcile(ctx vzspi.ComponentContext, mlc *modulesv1alpha1.ModuleLifecycle) error {
	log := ctx.Log()
	// Initialize the module if this is the first time we are reconciling it
	if err := initializeModule(ctx, mlc); err != nil {
		return err
	}
	condition := mlc.Status.Conditions[len(mlc.Status.Conditions)-1].Type
	switch condition {
	case modulesv1alpha1.CondPreInstall:
		return r.handlePreInstall(ctx, mlc, log)
	case modulesv1alpha1.CondInstallStarted:
		return r.handleInstallStarted(ctx, mlc, log)
	case modulesv1alpha1.CondPreUpgrade:
		return r.handlePreUpgrade(ctx, mlc, log)
	case modulesv1alpha1.CondInstallComplete, modulesv1alpha1.CondUpgradeComplete:
		return r.ReadyState(ctx, mlc)
	case modulesv1alpha1.CondUpgradeStarted:
		return r.handleUpgradeStarted(ctx, mlc, log)
	}
	return nil
}

func (r *Reconciler) handleUpgradeStarted(ctx vzspi.ComponentContext, mlc *modulesv1alpha1.ModuleLifecycle, log vzlog.VerrazzanoLogger) error {
	if r.comp.IsReady(ctx) {
		log.Progressf("Post-upgrade for %s is running", common.GetNamespacedName(mlc.ObjectMeta))
		if err := r.comp.PostUpgrade(ctx); err != nil {
			return err
		}
		mlc.Status.ObservedGeneration = mlc.Generation
		return UpdateStatus(ctx.Client(), mlc, string(modulesv1alpha1.CondUpgradeComplete), modulesv1alpha1.CondUpgradeComplete)
	}
	return delegates.NotReadyErrorf("Upgrade for %s is not ready", common.GetNamespacedName(mlc.ObjectMeta))
}

func (r *Reconciler) handlePreInstall(ctx vzspi.ComponentContext, mlc *modulesv1alpha1.ModuleLifecycle, log vzlog.VerrazzanoLogger) error {
	log.Progressf("Pre-install for %s is running", common.GetNamespacedName(mlc.ObjectMeta))
	if err := r.comp.PreInstall(ctx); err != nil {
		return err
	}
	if err := r.comp.Install(ctx); err != nil {
		return err
	}
	return UpdateStatus(ctx.Client(), mlc, string(modulesv1alpha1.CondInstallStarted), modulesv1alpha1.CondInstallStarted)
}

func (r *Reconciler) handlePreUpgrade(ctx vzspi.ComponentContext, mlc *modulesv1alpha1.ModuleLifecycle, log vzlog.VerrazzanoLogger) error {
	log.Progressf("Pre-upgrade for %s is running", common.GetNamespacedName(mlc.ObjectMeta))
	if err := r.comp.PreUpgrade(ctx); err != nil {
		return err
	}
	if err := r.comp.Upgrade(ctx); err != nil {
		return err
	}
	return UpdateStatus(ctx.Client(), mlc, string(modulesv1alpha1.CondUpgradeStarted), modulesv1alpha1.CondUpgradeStarted)
}

func (r *Reconciler) handleInstallStarted(ctx vzspi.ComponentContext, mlc *modulesv1alpha1.ModuleLifecycle, log vzlog.VerrazzanoLogger) error {
	if r.comp.IsReady(ctx) {
		log.Progressf("Post-install for %s is running", common.GetNamespacedName(mlc.ObjectMeta))
		if err := r.comp.PostInstall(ctx); err != nil {
			return err
		}
		mlc.Status.ObservedGeneration = mlc.Generation
		ctx.Log().Infof("%s is ready", common.GetNamespacedName(mlc.ObjectMeta))
		return UpdateStatus(ctx.Client(), mlc, string(modulesv1alpha1.CondInstallComplete), modulesv1alpha1.CondInstallComplete)
	}
	return delegates.NotReadyErrorf("Install for %s is not ready", common.GetNamespacedName(mlc.ObjectMeta))
}

// ReadyState reconciles put the Module back to pending state if the generation has changed
func (r *Reconciler) ReadyState(ctx vzspi.ComponentContext, mlc *modulesv1alpha1.ModuleLifecycle) error {
	if needsReconcile(mlc) {
		return UpdateStatus(ctx.Client(), mlc, string(modulesv1alpha1.CondPreUpgrade), modulesv1alpha1.CondPreUpgrade)
	}
	return nil
}

// Uninstall cleans up the component and removes the Module finalizer so Kubernetes can clean the resource
func (r *Reconciler) Uninstall(ctx vzspi.ComponentContext) error {
	if err := r.comp.PreUninstall(ctx); err != nil {
		return err
	}
	if err := r.comp.Uninstall(ctx); err != nil {
		return err
	}
	if err := r.comp.PostUninstall(ctx); err != nil {
		return err
	}
	return nil
}

func initializeModule(ctx vzspi.ComponentContext, mlc *modulesv1alpha1.ModuleLifecycle) error {
	initializeModuleStatus(ctx, mlc)
	return nil
}

func initializeModuleStatus(ctx vzspi.ComponentContext, mlc *modulesv1alpha1.ModuleLifecycle) {
	if len(mlc.Status.State) == 0 {
		mlc.SetState(modulesv1alpha1.StatePreinstall)
		mlc.Status.Conditions = []modulesv1alpha1.ModuleLifecycleCondition{
			NewCondition(string(modulesv1alpha1.StatePreinstall), modulesv1alpha1.CondPreInstall),
		}
	}
}
