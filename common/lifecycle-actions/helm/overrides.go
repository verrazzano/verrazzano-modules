// Copyright (c) 2022, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package helm

import (
	"fmt"
	moduleplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano/pkg/bom"
	"github.com/verrazzano/verrazzano/pkg/helm"
	vzoverride "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/common/override"
	"os"

	vzos "github.com/verrazzano/verrazzano/pkg/os"

	"github.com/verrazzano/verrazzano/platform-operator/apis/verrazzano/v1alpha1"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"

	"sigs.k8s.io/yaml"
)

// BuildCustomHelmOverrides Builds the helm overrides for a release, including image and file, and custom overrides
// - returns an error and a HelmOverride struct with the field populated
func BuildCustomHelmOverrides(context spi.ComponentContext, namespace string, moduleOverrides moduleplatform.Overrides) ([]helm.HelmOverrides, error) {
	// Optionally create a second override file.  This will contain both image setOverrides and any additional
	// setOverrides required by a component.
	// Get image setOverrides unless opt out
	var kvs []bom.KeyValue
	var err error
	var helmOverrides []helm.HelmOverrides

	vzOverrides := v1alpha1.Overrides{
		ConfigMapRef: moduleOverrides.ConfigMapRef,
		SecretRef:    moduleOverrides.SecretRef,
		Values:       moduleOverrides.Values,
	}

	// Getting user defined Helm overrides as the highest priority
	overrideStrings, err := vzoverride.GetInstallOverridesYAML(context, []v1alpha1.Overrides{vzOverrides})
	if err != nil {
		return nil, err
	}
	for _, overrideString := range overrideStrings {
		file, err := vzos.CreateTempFile(fmt.Sprintf("helm-overrides-user-%s-*.yaml", h.Name()), []byte(overrideString))
		if err != nil {
			context.Log().Error(err.Error())
			return nil, err
		}
		kvs = append(kvs, bom.KeyValue{Value: file.Name(), IsFile: true})
	}

	// Create files from the Verrazzano Helm values
	newKvs, err := filesFromVerrazzanoHelm(context, namespace, additionalValues)
	if err != nil {
		return overrides, err
	}
	kvs = append(kvs, newKvs...)

	// Add the values file ot the file overrides
	if len(h.ValuesFile) > 0 {
		kvs = append(kvs, bom.KeyValue{Value: h.ValuesFile, IsFile: true})
	}

	// Convert the key value pairs to Helm overrides
	overrides = h.organizeHelmOverrides(kvs)
	return overrides, nil
}

func filesFromVerrazzanoHelm(context spi.ComponentContext, namespace string, additionalValues []bom.KeyValue) ([]bom.KeyValue, error) {
	var kvs []bom.KeyValue
	var newKvs []bom.KeyValue

	// Get image overrides if they are specified
	imageOverrides, err := getImageOverrides(h.ReleaseName)
	if err != nil {
		return newKvs, err
	}
	kvs = append(kvs, imageOverrides...)

	// Append any additional setOverrides for the component (see Keycloak.go for example)
	if h.AppendOverridesFunc != nil {
		overrideValues, err := h.AppendOverridesFunc(context, h.ReleaseName, namespace, h.ChartDir, []bom.KeyValue{})
		if err != nil {
			return newKvs, err
		}
		kvs = append(kvs, overrideValues...)
	}

	// Append any special overrides passed in
	if len(additionalValues) > 0 {
		kvs = append(kvs, additionalValues...)
	}

	// Expand the existing kvs values into expected format
	var fileValues []bom.KeyValue
	for _, kv := range kvs {
		// If the value is a file, add it to the new kvs
		if kv.IsFile {
			newKvs = append(newKvs, kv)
			continue
		}

		// If set file, extract the data into the value parameter
		if kv.SetFile {
			data, err := os.ReadFile(kv.Value)
			if err != nil {
				return newKvs, context.Log().ErrorfNewErr("Could not open file %s: %v", kv.Value, err)
			}
			kv.Value = string(data)
		}

		fileValues = append(fileValues, kv)
	}

	// Take the YAML values and construct a YAML file
	// This uses the Helm YAML formatting
	fileString, err := yaml.HelmValueFileConstructor(fileValues)
	if err != nil {
		return newKvs, context.Log().ErrorfNewErr("Could not create YAML file from key value pairs: %v", err)
	}

	// Create the file from the string
	if len(fileString) > 0 {
		file, err := vzos.CreateTempFile(fmt.Sprintf("helm-overrides-verrazzano-%s-*.yaml", h.Name()), []byte(fileString))
		if err != nil {
			context.Log().Error(err.Error())
			return newKvs, err
		}
		if file != nil {
			newKvs = append(newKvs, bom.KeyValue{Value: file.Name(), IsFile: true})
		}
	}
	return newKvs, nil
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
