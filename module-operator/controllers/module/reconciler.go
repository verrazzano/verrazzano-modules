// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package module

import (
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/spi"
	compspi "github.com/verrazzano/verrazzano-modules/common/lifecycle-actions/action_spi"
	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	moduleplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"

	vzspi "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
)

const (
	defaultRepoName = "vz-stable"
	defaultRepoURI  = "http://localhost:8080"
)

var (
	trueValue = true
)

// Reconcile reconciles the Module CR
func (r Reconciler) Reconcile(spictx spi.ReconcileContext, u *unstructured.Unstructured) (ctrl.Result, error) {
	cr := &moduleplatform.Module{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, cr); err != nil {
		return ctrl.Result{}, err
	}
	ctx, err := vzspi.NewMinimalContext(r.Client, spictx.Log)
	if err != nil {
		return util.NewRequeueWithShortDelay(), err
	}

	helmInfo := loadHelmInfo(cr)
	tracker := getTracker(cr.ObjectMeta, stateInit)

	action := r.getAction(cr)
	smc := stateMachineContext{
		cr:        cr,
		tracker:   tracker,
		chartInfo: &helmInfo,
		action:    action,
	}

	res := r.doStateMachine(ctx, smc)
	return res, nil
}

func loadHelmInfo(cr *moduleplatform.Module) compspi.HelmInfo {
	helmInfo := compspi.HelmInfo{
		HelmRelease: &moduleplatform.HelmRelease{
			Name:      "vz-integration-operator",
			Namespace: "default",
			ChartInfo: moduleplatform.HelmChart{
				Name:    "vz-integration-operator",
				Version: "0.1.0",
				Path:    "/Users/pmackin/charts/vz-integration-operator-0.1.0.tgz",
			},
			Overrides: nil,
		},
	}
	return helmInfo
}

func (r *Reconciler) getAction(cr *moduleplatform.Module) compspi.LifecycleActionHandler {
	return r.comp.InstallAction
}

func (r *Reconciler) lookupModuleSource(mod *moduleplatform.Module) (repoName, sourceURI string) {
	source := mod.Spec.Source
	if source == nil {
		return defaultRepoName, defaultRepoURI
	}
	return source.ChartRepo.Name, source.ChartRepo.URI
}

func (r *Reconciler) lookupChartNamespace(mod *moduleplatform.Module) string {
	if len(mod.Spec.TargetNamespace) > 0 {
		return mod.Spec.TargetNamespace
	}
	return mod.Namespace
}

func (r *Reconciler) lookupChartName(moduleInstance *moduleplatform.Module) string {
	if len(moduleInstance.Spec.Name) > 0 {
		return moduleInstance.Spec.Name
	}
	return moduleInstance.Name
}
