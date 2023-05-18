// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package common

import (
	"github.com/verrazzano/verrazzano-modules/common/handlerspi"
	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	"github.com/verrazzano/verrazzano-modules/common/pkg/helm"
	"github.com/verrazzano/verrazzano-modules/common/pkg/semver"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"helm.sh/helm/v3/pkg/release"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ handlerspi.ModuleActualStateInCluster = &moduleState{}

type moduleState struct{}

// GetActualModuleState gets the state of the module
func (m moduleState) GetActualModuleState(context handlerspi.HandlerContext, cr *moduleapi.ModuleLifecycle) (handlerspi.ModuleActualState, ctrl.Result, error) {
	releaseName := cr.Spec.Installer.HelmRelease.Name
	releaseNamespace := cr.Spec.Installer.HelmRelease.Namespace
	releaseStatus, err := helm.GetReleaseStatus(context.Log, releaseName, releaseNamespace)
	if err != nil {
		context.Log.ErrorfThrottled("Failed getting Helm release %s/%s failed with error: %v\n", releaseNamespace, releaseName, err)
		return handlerspi.ModuleStateUnknown, util.NewRequeueWithShortDelay(), err
	}
	switch release.Status(releaseStatus) {
	case release.StatusUnknown:
		return handlerspi.ModuleStateNotInstalled, ctrl.Result{}, nil
	case release.StatusDeployed:
		return handlerspi.ModuleStateReady, ctrl.Result{}, nil
	case release.StatusFailed:
		return handlerspi.ModuleStateFailed, ctrl.Result{}, nil
	default:
		return handlerspi.ModuleStateReconciling, ctrl.Result{}, nil
	}
}

// IsUpgradeNeeded checks if upgrade is needed
func (m moduleState) IsUpgradeNeeded(context handlerspi.HandlerContext, cr *moduleapi.ModuleLifecycle) (bool, ctrl.Result, error) {
	releaseName := cr.Spec.Installer.HelmRelease.Name
	releaseNamespace := cr.Spec.Installer.HelmRelease.Namespace
	installedVersion, err := helm.GetReleaseChartVersion(releaseName, releaseNamespace)
	if err != nil {
		context.Log.ErrorfThrottled("Failed getting version for Helm release %s/%s failed with error: %v\n", releaseNamespace, releaseName, err)
		return false, util.NewRequeueWithShortDelay(), err
	}

	// return UpgradeAction only when the desired version is different from current
	upgradeNeeded, err := IsUpgradeNeeded(cr.Spec.Version, installedVersion)
	if err != nil {
		context.Log.ErrorfThrottled("Failed checking if upgrade needed for Helm release %s/%s failed with error: %v\n", releaseNamespace, releaseName, err)
		return false, util.NewRequeueWithShortDelay(), err
	}
	return upgradeNeeded, ctrl.Result{}, nil
}

// IsUpgradeNeeded returns true if upgrade is needed
func IsUpgradeNeeded(desiredVersion, installedVersion string) (bool, error) {
	desiredSemver, err := semver.NewSemVersion(desiredVersion)
	if err != nil {
		return false, err
	}
	installedSemver, err := semver.NewSemVersion(installedVersion)
	if err != nil {
		return false, err
	}
	return installedSemver.IsLessThan(desiredSemver), nil
}
