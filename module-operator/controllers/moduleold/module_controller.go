// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package moduleold

import (
	"context"
	"github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/helm"
	"github.com/verrazzano/verrazzano-modules/common/pkg/k8s"
	"time"

	modulesv1alpha1 "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	vzcontroller "github.com/verrazzano/verrazzano/pkg/controller"
	"github.com/verrazzano/verrazzano/pkg/log/vzlog"
	vzstring "github.com/verrazzano/verrazzano/pkg/string"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciler reconciles a Verrazzano Platform object
type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Controller controller.Controller
}

const (
	// Name of finalizer
	finalizerName = "modules.verrazzano.io"

	defaultRepoName = "vz-stable"
	defaultRepoURI  = "http://localhost:8080"
)

var (
	trueValue = true
)

// SetupWithManager creates a new controller and adds it to the manager
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	var err error
	r.Controller, err = ctrl.NewControllerManagedBy(mgr).
		For(&modulesv1alpha1.Module{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 10,
		}).
		Build(r)
	return err
}

// Reconcile the Module CR
// +kubebuilder:rbac:groups=install.verrazzano.io,resources=modules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=install.verrazzano.io,resources=modules/status,verbs=get;update;patch
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	// NOTE: Metrics setup

	moduleInstance := &modulesv1alpha1.Module{}
	if err := r.Get(ctx, req.NamespacedName, moduleInstance); err != nil {
		// NOTE: errorCounterMetricObject.Inc()
		// If the resource is not found, that means all the finalizers have been removed,
		// and the Verrazzano resource has been deleted, so there is nothing left to do.
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		zap.S().Errorf("Failed to fetch Module resource: %v", err)
		return newRequeueWithDelay(), nil
	}

	// Get the resource logger
	log, err := vzlog.EnsureResourceLogger(&vzlog.ResourceConfig{
		Name:           moduleInstance.Name,
		Namespace:      moduleInstance.Namespace,
		ID:             string(moduleInstance.UID),
		Generation:     moduleInstance.Generation,
		ControllerName: "vzmodule",
	})
	if err != nil {
		// NOTE: errorCounterMetricObject.Inc()
		zap.S().Errorf("Failed to create controller logger for Module controller: %v", err)
	}

	// Check if resource is being deleted
	if !moduleInstance.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Oncef("Removing finalizer %s", finalizerName)
		moduleInstance.ObjectMeta.Finalizers = vzstring.RemoveStringFromSlice(moduleInstance.ObjectMeta.Finalizers, finalizerName)
		if err := r.Update(ctx, moduleInstance); err != nil {
			return newRequeueWithDelay(), err
		}
		return ctrl.Result{}, nil
	}

	if !vzstring.SliceContainsString(moduleInstance.ObjectMeta.Finalizers, finalizerName) {
		log.Debugf("Adding finalizer %s", finalizerName)
		moduleInstance.ObjectMeta.Finalizers = append(moduleInstance.ObjectMeta.Finalizers, finalizerName)
		if err := r.Update(context.TODO(), moduleInstance); err != nil {
			return newRequeueWithDelay(), err
		}
	}

	return r.doReconcile(log, moduleInstance)
}

func (r *Reconciler) doReconcile(log vzlog.VerrazzanoLogger, moduleInstance *modulesv1alpha1.Module) (ctrl.Result, error) {
	log.Infof("Reconciling Verrazzano module instance %s/%s", moduleInstance.Namespace, moduleInstance.Name)

	sourceName, sourceURI := r.lookupModuleSource(moduleInstance)

	chartName := r.lookupChartName(moduleInstance)
	chartNamespace := r.lookupChartNamespace(moduleInstance)

	// Find the desired module version
	targetModuleVersion, err := r.lookupModuleVersion(log, moduleInstance, chartName, sourceName, sourceURI)
	if err != nil {
		return vzcontroller.NewRequeueWithDelay(5, 10, time.Second), err
	}

	if _, err := r.reconcileModule(moduleInstance, chartName, chartNamespace, targetModuleVersion, sourceName, sourceURI); err != nil {
		return newRequeueWithDelay(), err
	}
	if moduleInstance.Status.State != modulesv1alpha1.ModuleStateReady {
		// Not in a ready state yet, requeue and re-check
		log.Progressf("Module %s reconciling, requeue", k8s.GetNamespacedName(moduleInstance.ObjectMeta))
		return newRequeueWithDelay(), nil
	}
	log.Infof("Module %s/%s reconcile complete", moduleInstance.Namespace, moduleInstance.Name)
	return ctrl.Result{}, nil
}

func (r *Reconciler) reconcileModule(mod *modulesv1alpha1.Module, chartName string, chartNamespace string, moduleVersion string, sourceName string, sourceURI string) (*modulesv1alpha1.ModuleLifecycle, error) {
	lifecycleResource, err := r.createLifecycleResource(sourceName, sourceURI, chartName, chartNamespace, moduleVersion,
		modulesv1alpha1.Overrides{}, createOwnerRef(mod))
	if err != nil {
		return nil, err
	}
	if err := r.updateModuleInstanceState(mod, lifecycleResource); err != nil {
		return nil, err
	}
	return lifecycleResource, err
}

func (r *Reconciler) createLifecycleResource(sourceName string, sourceURI string, chartName string, chartNamespace string, chartVersion string, overrides modulesv1alpha1.Overrides, ownerRef *metav1.OwnerReference) (*modulesv1alpha1.ModuleLifecycle, error) {

	// Create a CR to manage the module installation
	moduleInstaller := &modulesv1alpha1.ModuleLifecycle{
		ObjectMeta: metav1.ObjectMeta{
			Name:      chartName,
			Namespace: chartNamespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, moduleInstaller, func() error {
		if moduleInstaller.ObjectMeta.Labels == nil {
			moduleInstaller.ObjectMeta.Labels = make(map[string]string)
		}
		moduleInstaller.Spec = modulesv1alpha1.ModuleLifecycleSpec{
			LifecycleClassName: modulesv1alpha1.HelmLifecycleClass,
			Installer: modulesv1alpha1.ModuleInstaller{
				HelmRelease: &modulesv1alpha1.HelmRelease{
					Name:      chartName, // REVIEW: should this be associated with the Module name?
					Namespace: chartNamespace,
					Repository: modulesv1alpha1.HelmChartRepository{
						Name: sourceName,
						URI:  sourceURI,
					},
					ChartInfo: modulesv1alpha1.HelmChart{
						Name:    chartName,
						Version: chartVersion,
					},
					Overrides: []modulesv1alpha1.Overrides{overrides},
				},
			},
		}
		if ownerRef != nil {
			if !ownerRefExists(moduleInstaller, ownerRef) {
				moduleInstaller.OwnerReferences = append(moduleInstaller.OwnerReferences, *ownerRef)
			}
		}
		return nil
	})
	return moduleInstaller, err
}

func ownerRefExists(moduleInstaller *modulesv1alpha1.ModuleLifecycle, ownerRef *metav1.OwnerReference) bool {
	for _, ref := range moduleInstaller.OwnerReferences {
		if ref.UID == ownerRef.UID {
			return true
		}
	}
	return false
}

func (r *Reconciler) lookupModuleSource(mod *modulesv1alpha1.Module) (repoName, sourceURI string) {
	source := mod.Spec.Source
	if source == nil {
		return defaultRepoName, defaultRepoURI
	}
	return source.ChartRepo.Name, source.ChartRepo.URI
}

func (r *Reconciler) lookupChartNamespace(mod *modulesv1alpha1.Module) string {
	if len(mod.Spec.TargetNamespace) > 0 {
		return mod.Spec.TargetNamespace
	}
	return mod.Namespace
}

func (r *Reconciler) lookupChartName(moduleInstance *modulesv1alpha1.Module) string {
	if len(moduleInstance.Spec.Name) > 0 {
		return moduleInstance.Spec.Name
	}
	return moduleInstance.Name
}

func (r *Reconciler) lookupModuleVersion(log vzlog.VerrazzanoLogger, moduleInstance *modulesv1alpha1.Module, chartName string, repoName string, repoURI string) (string, error) {
	// Find target module version
	// - declared in the Module instance
	var modVersion string
	// Look up the explicitly declared module version
	if len(moduleInstance.Spec.Version) > 0 {
		return moduleInstance.Spec.Version, nil
	}
	// - find the most recent module version in the repo
	modVersion, err := helm.FindLatestChartVersion(log, chartName, repoName, repoURI)
	if err != nil {
		return "", err
	}

	return modVersion, nil
}

func (r *Reconciler) updateModuleInstanceState(instance *modulesv1alpha1.Module, lifecycleResource *modulesv1alpha1.ModuleLifecycle) error {
	installerState := lifecycleResource.Status.State
	switch installerState {
	case modulesv1alpha1.StateReady:
		instance.Status.State = modulesv1alpha1.ModuleStateReady
		helmRelease := lifecycleResource.Spec.Installer.HelmRelease
		if helmRelease != nil {
			instance.Status.Version = helmRelease.ChartInfo.Version
		}
	default:
		instance.Status.State = modulesv1alpha1.ModuleStateReconciling
	}
	return r.Status().Update(context.TODO(), instance)
}

func createOwnerRef(owner *modulesv1alpha1.Module) *metav1.OwnerReference {
	return &metav1.OwnerReference{
		APIVersion:         owner.APIVersion,
		Kind:               owner.Kind,
		Name:               owner.Name,
		UID:                owner.UID,
		Controller:         &trueValue,
		BlockOwnerDeletion: &trueValue,
	}
}

func newRequeueWithDelay() ctrl.Result {
	return vzcontroller.NewRequeueWithDelay(2, 5, time.Second)
}
