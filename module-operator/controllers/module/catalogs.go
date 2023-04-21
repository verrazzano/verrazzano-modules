// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package module

import (
	"github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
)

type ModuleCatalog struct {
	Version            string                        `json:"version"`
	LifecycleOperators []LifecycleOperatorDefinition `json:"lifecycleOperators,omitempty"`
	ModuleDefinitions  []ModuleDefinition            `json:"moduleDefinitions,omitempty"`
}

// LifecycleOperatorDefinition specifies a metadata about an operator chart type.
type LifecycleOperatorDefinition struct {
	LifecycleClassName string             `json:"lifecycleClassname,omitempty"`
	Chart              v1alpha1.HelmChart `json:"chart,omitempty"`
}

// ModuleDefinition defines properties of a Module chart type
type ModuleDefinition struct {
	LifecycleOperatorDefinition `json:",inline"`
	Dependencies                []ChartDependency `json:"dependencies,omitempty"`
}

type ChartDependency struct {
	Name              string `json:"name"`
	Version           string `json:"version,omitempty"`
	SupportedVersions string `json:"supportedVersions,omitempty"`
	Chart             v1alpha1.HelmChart
}
