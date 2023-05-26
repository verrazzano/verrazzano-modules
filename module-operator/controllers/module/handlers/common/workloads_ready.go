// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package common

import (
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/handlerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/helm"
	"github.com/verrazzano/verrazzano-modules/pkg/k8s/readiness"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"
	"strings"
)

func CheckWorkLoadsReady(ctx handlerspi.HandlerContext, releaseName string, namespace string) (bool, error) {
	const yamlSep = "---"

	type Resource struct {
		Kind     string `json:"kind"`
		Metadata struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
		}
	}

	// Get all the deployments, statefulsets, and daemonsets for this Helm release
	rel, err := helm.GetRelease(ctx.Log, releaseName, namespace)
	if err != nil {
		return false, err
	}
	if rel.Manifest == "" {
		return false, err
	}

	// Get the manifests objects
	resYamls := strings.Split(rel.Manifest, yamlSep)
	for _, yam := range resYamls {
		if strings.TrimSpace(yam) == "" {
			continue
		}
		res := Resource{}
		if err := yaml.Unmarshal([]byte(yam), &res); err != nil {
			return false, err
		}
		nsns := []types.NamespacedName{{Namespace: res.Metadata.Namespace, Name: res.Metadata.Name}}
		switch strings.ToLower(res.Kind) {
		case "deployment":
			if ready := readiness.DeploymentsAreReady(ctx.Log, ctx.Client, nsns, releaseName); !ready {
				return false, nil
			}
		case "statefuleset":
			if ready := readiness.StatefulSetsAreReady(ctx.Log, ctx.Client, nsns, releaseName); !ready {
				return false, nil
			}
		case "daemonset":
			if ready := readiness.DaemonSetsAreReady(ctx.Log, ctx.Client, nsns, releaseName); !ready {
				return false, nil
			}
		}
	}
	return true, err
}
