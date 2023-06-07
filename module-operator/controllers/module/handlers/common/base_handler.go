// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package common

import (
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/handlerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/util"
	helm2 "github.com/verrazzano/verrazzano-modules/pkg/helm"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	"helm.sh/helm/v3/pkg/release"
	ctrl "sigs.k8s.io/controller-runtime"
)

// upgradeFuncSig is a function needed for unit test override
type upgradeFuncSig func(log vzlog.VerrazzanoLogger, releaseOpts *helm2.HelmReleaseOpts, wait bool, dryRun bool) (*release.Release, error)

var upgradeFunc upgradeFuncSig = helm2.UpgradeRelease

type BaseHandler struct{}

func SetUpgradeFunc(f upgradeFuncSig) {
	upgradeFunc = f
}

func ResetUpgradeFunc() {
	upgradeFunc = helm2.UpgradeRelease
}

// HelmUpgradeOrInstall does a Helm upgrade --install of the chart
func (h BaseHandler) HelmUpgradeOrInstall(ctx handlerspi.HandlerContext) (ctrl.Result, error) {
	module := ctx.CR.(*moduleapi.Module)

	// Perform a Helm install using the helm upgrade --install command
	helmRelease := ctx.HelmInfo.HelmRelease
	helmOverrides, err := helm2.LoadOverrideFiles(ctx.Log, ctx.Client, helmRelease.Name, module.Namespace, buildOverrides(module))
	if err != nil {
		return ctrl.Result{}, err
	}
	var opts = &helm2.HelmReleaseOpts{
		RepoURL:      helmRelease.Repository.URI,
		ReleaseName:  helmRelease.Name,
		Namespace:    helmRelease.Namespace,
		ChartPath:    helmRelease.ChartInfo.Path,
		ChartVersion: helmRelease.ChartInfo.Version,
		Overrides:    helmOverrides,
	}
	_, err = upgradeFunc(ctx.Log, opts, false, ctx.DryRun)
	return ctrl.Result{}, err
}

// CheckReleaseDeployedAndReady checks if the Helm release is deployed and ready
func (h BaseHandler) CheckReleaseDeployedAndReady(ctx handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	if ctx.DryRun {
		ctx.Log.Debugf("IsReady() dry run for %s", ctx.HelmRelease.Name)
		return true, ctrl.Result{}, nil
	}
	// Check if the Helm release is deployed
	deployed, err := helm2.IsReleaseDeployed(ctx.HelmRelease.Name, ctx.HelmRelease.Namespace)
	if err != nil {
		ctx.Log.ErrorfThrottled("Error occurred checking release deployment: %v", err.Error())
		return false, ctrl.Result{}, err
	}
	if !deployed {
		return false, util.NewRequeueWithShortDelay(), nil
	}

	// Check if the workload pods are ready
	ready, err := CheckWorkLoadsReady(ctx, ctx.HelmRelease.Name, ctx.HelmRelease.Namespace)
	return ready, ctrl.Result{}, err
}

// buildOverrides builds the Helm value overrides in the correct precedence order, where values has the highest precedence.
func buildOverrides(module *moduleapi.Module) []helm2.ValueOverrides {
	overrides := []helm2.ValueOverrides{}

	// Add all the valueFrom overrides
	for i := range module.Spec.ValuesFrom {
		// Skip creating an entry if both refs are nil
		if module.Spec.ValuesFrom[i].ConfigMapRef == nil && module.Spec.ValuesFrom[i].SecretRef == nil {
			continue
		}
		override := helm2.ValueOverrides{
			ConfigMapRef: module.Spec.ValuesFrom[i].ConfigMapRef,
			SecretRef:    module.Spec.ValuesFrom[i].SecretRef,
		}
		overrides = append(overrides, override)
	}

	// Add the values overrides last, so they have the highest precedence
	if module.Spec.Values != nil {
		override := helm2.ValueOverrides{
			Values: module.Spec.Values,
		}
		overrides = append(overrides, override)
	}

	return overrides
}
