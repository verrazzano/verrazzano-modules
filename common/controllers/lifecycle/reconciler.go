// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package lifecycle

import (
	compspi "github.com/verrazzano/verrazzano-modules/common/component/spi"
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/spi"
	"github.com/verrazzano/verrazzano-modules/common/k8s"
	helmcomp "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/helm"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"

	modulesv1alpha1 "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/common"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/modlifecycle/delegates"

	modplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"

	vzctrl "github.com/verrazzano/verrazzano-modules/module-operator/pkg/controller"
	"github.com/verrazzano/verrazzano/pkg/log/vzlog"
	vzconst "github.com/verrazzano/verrazzano/platform-operator/constants"
	vzspi "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
)

// componentInstallState identifies the state of a component during install
type componentInstallState string

const (
	// compStateInstallInitDetermineComponentState is the state when a component is initialized
	compStateInstallInit componentInstallState = "componentStateInit"

	// compStateInstallingUpdate is the state when the status is updated to installing
	compStateInstallingUpdate componentInstallState = "compStateInstallingUpdate"

	// compStatePreInstall is the state when a component does a pre-install
	compStatePreInstall componentInstallState = "compStatePreInstall"

	// compStateInstall is the state where a component does an install
	compStateInstall componentInstallState = "compStateInstall"

	// compStateInstallWaitReady is the state when a component is waiting for install to be ready
	compStateInstallWaitReady componentInstallState = "compStateInstallWaitReady"

	// compStatePostInstall is the state when a component is doing a post-install
	compStatePostInstall componentInstallState = "compStatePostInstall"

	// compStateInstallCompleteUpdate is the state when component writes the Install Complete status
	compStateInstallCompleteUpdate componentInstallState = "compStateInstallCompleteUpdate"

	// compStateInstallEnd is the terminal state
	compStateInstallEnd componentInstallState = "compStateInstallEnd"
)

// componentTrackerContext has the component context tracker
type componentTrackerContext struct {
	installState componentInstallState
}

// Reconcile updates the Certificate
func (r Reconciler) Reconcile(spictx spi.ReconcileContext, u *unstructured.Unstructured) (ctrl.Result, error) {
	mlc := &modplatform.ModuleLifecycle{}
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

	helmInfo := loadHelmInfo(mlc)
	tracker := getInstallTracker(mlc.ObjectMeta, string(compStateInstallInit))

	res, err := r.doStateMachine(ctx, mlc, tracker, helmInfo)
	if err != nil {
		return newRequeueWithDelay(), err
	}
	if vzctrl.ShouldRequeue(res) {
		return res, nil
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) doStateMachine(spiCtx vzspi.ComponentContext, mlc *modplatform.ModuleLifecycle, tracker *installTracker, chartInfo compspi.HelmInfo) (ctrl.Result, error) {
	compName := common.GetNamespacedNameString(mlc.ObjectMeta)
	compContext := spiCtx.Init("component").Operation(vzconst.InstallOperation)
	compLog := compContext.Log()

	for tracker.state != string(compStateInstallEnd) {
		switch componentInstallState(tracker.state) {
		case compStateInstallInit:
			if err := r.comp.Init(compContext, &chartInfo); err != nil {
				return ctrl.Result{}, err
			}
			tracker.state = string(compStateInstallingUpdate)

		case compStateInstallingUpdate:
			if err := UpdateStatus(r.Client, mlc, string(modulesv1alpha1.CondInstallStarted), modulesv1alpha1.CondInstallStarted); err != nil {
				return ctrl.Result{}, err
			}
			tracker.state = string(compStatePreInstall)

		case compStatePreInstall:
			if err := r.comp.PreInstall(compContext); err != nil {
				return ctrl.Result{}, err
			}
			tracker.state = string(compStateInstall)

		case compStateInstall:
			if err := r.comp.Install(compContext); err != nil {
				return ctrl.Result{}, err
			}
			tracker.state = string(compStateInstallWaitReady)

		case compStateInstallWaitReady:
			if !r.comp.IsReady(compContext) {
				compLog.Progressf("Component %s has been installed. Waiting for the component to be ready", compName)
				return newRequeueWithDelay(), nil
			}
			compLog.Oncef("Component %s successfully installed and is ready", r.comp.Name())
			tracker.state = string(compStatePostInstall)

		case compStatePostInstall:
			compLog.Oncef("Component %s post-install running", compName)
			if err := r.comp.PostInstall(compContext); err != nil {
				return ctrl.Result{}, err
			}
			tracker.state = string(compStateInstallCompleteUpdate)

		case compStateInstallCompleteUpdate:
			if err := UpdateStatus(r.Client, mlc, string(modulesv1alpha1.CondUpgradeComplete), modulesv1alpha1.CondUpgradeComplete); err != nil {
				return ctrl.Result{}, err
			}
			tracker.state = string(compStateInstallEnd)
		}
	}
	return ctrl.Result{}, nil
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

func newRequeueWithDelay() ctrl.Result {
	return vzctrl.NewRequeueWithDelay(3, 10, time.Second)
}

func loadHelmInfo(mlc *modplatform.ModuleLifecycle) compspi.HelmInfo {
	chartInfo := compspi.HelmInfo{
		HelmComponent: helmcomp.HelmComponent{},
		HelmRelease:   mlc.Spec.Installer.HelmRelease,
	}

	return chartInfo
}
