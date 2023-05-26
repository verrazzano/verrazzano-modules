// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package common

import (
	"context"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/handlerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/helm"
	"github.com/verrazzano/verrazzano-modules/pkg/k8s/readiness"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const helmKey = " meta.helm.sh/release-name"

func CheckWorkLoadsReady(ctx handlerspi.HandlerContext, releaseName string, namespace string) (bool, error) {
	// Get all the deployments, statefulsets, and daemonsets for this Helm release
	rel, err := helm.GetRelease(ctx.Log, releaseName, namespace)
	if err != nil {
		return false, err
	}
	if rel.Manifest == "" {
		return false, err
	}
	ready := checkDeploymentsReady(ctx, releaseName, namespace) && checkStatefulSetsReady(ctx, releaseName, namespace)
	checkDaemonSetsReady(ctx, releaseName, namespace)

	return ready, nil
}

func checkDeploymentsReady(ctx handlerspi.HandlerContext, releaseName string, namespace string) bool {
	depList := v1.DeploymentList{}
	err := ctx.Client.List(context.TODO(), &depList, &client.ListOptions{
		Namespace: namespace,
	})
	if err != nil {
		return false
	}
	for _, dep := range depList.Items {
		if dep.Annotations == nil {
			continue
		}
		rel, ok := dep.Annotations[helmKey]
		if ok && rel == releaseName {
			nsns := []types.NamespacedName{{
				Namespace: dep.Namespace,
				Name:      dep.Name,
			}}
			if ready := readiness.DeploymentsAreReady(ctx.Log, ctx.Client, nsns, releaseName); !ready {
				return false
			}
		}
	}
	return true
}

func checkStatefulSetsReady(ctx handlerspi.HandlerContext, releaseName string, namespace string) bool {
	stsList := v1.StatefulSetList{}
	err := ctx.Client.List(context.TODO(), &stsList, &client.ListOptions{
		Namespace: namespace,
	})
	if err != nil {
		return false
	}
	for _, sts := range stsList.Items {
		if sts.Annotations == nil {
			continue
		}
		rel, ok := sts.Annotations[helmKey]
		if ok && rel == releaseName {
			nsns := []types.NamespacedName{{
				Namespace: sts.Namespace,
				Name:      sts.Name,
			}}
			if ready := readiness.StatefulSetsAreReady(ctx.Log, ctx.Client, nsns, releaseName); !ready {
				return false
			}
		}
	}
	return true
}

func checkDaemonSetsReady(ctx handlerspi.HandlerContext, releaseName string, namespace string) bool {
	demList := v1.DaemonSetList{}
	err := ctx.Client.List(context.TODO(), &demList, &client.ListOptions{
		Namespace: namespace,
	})
	if err != nil {
		return false
	}
	for _, dem := range demList.Items {
		if dem.Annotations == nil {
			continue
		}
		rel, ok := dem.Annotations[helmKey]
		if ok && rel == releaseName {
			nsns := []types.NamespacedName{{
				Namespace: dem.Namespace,
				Name:      dem.Name,
			}}
			if ready := readiness.DeploymentsAreReady(ctx.Log, ctx.Client, nsns, releaseName); !ready {
				return false
			}
		}
	}
	return true
}
