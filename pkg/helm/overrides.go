// Copyright (c) 2022, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package helm

import (
	"fmt"
	vzos "github.com/verrazzano/verrazzano-modules/pkg/os"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ValueOverrides identifies overrides for a Helm release.
type ValueOverrides struct {
	// Selector for ConfigMap containing override data.
	// +optional
	ConfigMapRef *corev1.ConfigMapKeySelector `json:"configMapRef,omitempty"`
	// Selector for Secret containing override data.
	// +optional
	SecretRef *corev1.SecretKeySelector `json:"secretRef,omitempty"`
	// Configure overrides using inline YAML.
	// +optional
	Values *apiextensionsv1.JSON `json:"values,omitempty"`
}

// LoadOverrideFiles loads the helm overrides into a set of files for a release.  Return a list of Helm overrides which contain the filenames
func LoadOverrideFiles(log vzlog.VerrazzanoLogger, client ctrlclient.Client, releaseName string, mlcNamespace string, moduleOverrides []ValueOverrides) ([]HelmOverrides, error) {
	if len(moduleOverrides) == 0 {
		return nil, nil
	}
	var kvs []KeyValue
	var err error

	// Getting user defined Helm overrides as the highest priority
	overrideStrings, err := GetInstallOverridesYAML(log, client, moduleOverrides, mlcNamespace)
	if err != nil {
		return nil, err
	}
	for _, overrideString := range overrideStrings {
		file, err := vzos.CreateTempFile(fmt.Sprintf("helm-overrides-release-%s-*.yaml", releaseName), []byte(overrideString))
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		kvs = append(kvs, KeyValue{Value: file.Name(), IsFile: true})
	}

	// Convert the key value pairs to Helm overrides
	overrides := organizeHelmOverrides(kvs)
	return overrides, nil
}

// organizeHelmOverrides creates a list of Helm overrides from key value pairs in reverse precedence (0th value has the lowest precedence)
// Each key value pair gets its own override object to keep strict precedence
func organizeHelmOverrides(kvs []KeyValue) []HelmOverrides {
	var overrides []HelmOverrides
	for _, kv := range kvs {
		if kv.SetString {
			// Append in reverse order because helm precedence is right to left
			overrides = append([]HelmOverrides{{SetStringOverrides: fmt.Sprintf("%s=%s", kv.Key, kv.Value)}}, overrides...)
		} else if kv.SetFile {
			// Append in reverse order because helm precedence is right to left
			overrides = append([]HelmOverrides{{SetFileOverrides: fmt.Sprintf("%s=%s", kv.Key, kv.Value)}}, overrides...)
		} else if kv.IsFile {
			// Append in reverse order because helm precedence is right to left
			overrides = append([]HelmOverrides{{FileOverride: kv.Value}}, overrides...)
		} else {
			// Append in reverse order because helm precedence is right to left
			overrides = append([]HelmOverrides{{SetOverrides: fmt.Sprintf("%s=%s", kv.Key, kv.Value)}}, overrides...)
		}
	}
	return overrides
}
