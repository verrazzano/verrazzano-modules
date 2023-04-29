// Copyright (c) 2022, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package helm

import (
	"fmt"
	moduleplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano/pkg/bom"
	"github.com/verrazzano/verrazzano/pkg/helm"
	vzos "github.com/verrazzano/verrazzano/pkg/os"
	"github.com/verrazzano/verrazzano/platform-operator/apis/verrazzano/v1beta1"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
)

// LoadOverrideFiles loads the helm overrides into a set of files for a release.  Return a list of Helm overrides which contain the filenames
func LoadOverrideFiles(context spi.ComponentContext, releaseName string, mlcNamespace string, moduleOverrides []moduleplatform.Overrides) ([]helm.HelmOverrides, error) {
	if len(moduleOverrides) == 0 {
		return nil, nil
	}
	var kvs []bom.KeyValue
	var err error
	vzOverrides := []v1beta1.Overrides{}

	for _, modOverride := range moduleOverrides {
		vzOverride := v1beta1.Overrides{
			ConfigMapRef: modOverride.ConfigMapRef,
			SecretRef:    modOverride.SecretRef,
			Values:       modOverride.Values,
		}
		vzOverrides = append(vzOverrides, vzOverride)
	}

	// Getting user defined Helm overrides as the highest priority
	overrideStrings, err := getInstallOverridesYAML(context.Log(), context.Client(), vzOverrides, mlcNamespace)
	if err != nil {
		return nil, err
	}
	for _, overrideString := range overrideStrings {
		file, err := vzos.CreateTempFile(fmt.Sprintf("helm-overrides-release-%s-*.yaml", releaseName), []byte(overrideString))
		if err != nil {
			context.Log().Error(err.Error())
			return nil, err
		}
		kvs = append(kvs, bom.KeyValue{Value: file.Name(), IsFile: true})
	}

	// Convert the key value pairs to Helm overrides
	overrides := organizeHelmOverrides(kvs)
	return overrides, nil
}

// organizeHelmOverrides creates a list of Helm overrides from key value pairs in reverse precedence (0th value has the lowest precedence)
// Each key value pair gets its own override object to keep strict precedence
func organizeHelmOverrides(kvs []bom.KeyValue) []helm.HelmOverrides {
	var overrides []helm.HelmOverrides
	for _, kv := range kvs {
		if kv.SetString {
			// Append in reverse order because helm precedence is right to left
			overrides = append([]helm.HelmOverrides{{SetStringOverrides: fmt.Sprintf("%s=%s", kv.Key, kv.Value)}}, overrides...)
		} else if kv.SetFile {
			// Append in reverse order because helm precedence is right to left
			overrides = append([]helm.HelmOverrides{{SetFileOverrides: fmt.Sprintf("%s=%s", kv.Key, kv.Value)}}, overrides...)
		} else if kv.IsFile {
			// Append in reverse order because helm precedence is right to left
			overrides = append([]helm.HelmOverrides{{FileOverride: kv.Value}}, overrides...)
		} else {
			// Append in reverse order because helm precedence is right to left
			overrides = append([]helm.HelmOverrides{{SetOverrides: fmt.Sprintf("%s=%s", kv.Key, kv.Value)}}, overrides...)
		}
	}
	return overrides
}