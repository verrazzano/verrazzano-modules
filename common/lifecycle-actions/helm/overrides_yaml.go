package helm

import (
	"context"

	"github.com/verrazzano/verrazzano/pkg/log/vzlog"
	"github.com/verrazzano/verrazzano/platform-operator/apis/verrazzano/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"
)

// getInstallOverridesYAML takes the list of Overrides and returns a string array of YAMLs
func getInstallOverridesYAML(log vzlog.VerrazzanoLogger, client client.Client, overrides []v1beta1.Overrides,
	mlcNamespace string) ([]string, error) {
	var overrideStrings []string
	for _, override := range overrides {
		// Check if ConfigMapRef is populated and gather data
		if override.ConfigMapRef != nil {
			// Get the ConfigMap data
			data, err := getConfigMapOverrides(log, client, override.ConfigMapRef, mlcNamespace)
			if err != nil {
				return overrideStrings, err
			}
			overrideStrings = append(overrideStrings, data)
			continue
		}
		// Check if SecretRef is populated and gather data
		if override.SecretRef != nil {
			// Get the Secret data
			data, err := getSecretOverrides(log, client, override.SecretRef, mlcNamespace)
			if err != nil {
				return overrideStrings, err
			}
			overrideStrings = append(overrideStrings, data)
			continue
		}
		if override.Values != nil {
			overrideValuesData, err := yaml.Marshal(override.Values)
			if err != nil {
				return overrideStrings, err
			}
			overrideStrings = append(overrideStrings, string(overrideValuesData))
		}

	}
	return overrideStrings, nil
}

// getConfigMapOverrides takes a ConfigMap selector and returns the YAML data and handles k8s api errors appropriately
func getConfigMapOverrides(log vzlog.VerrazzanoLogger, client client.Client, selector *v1.ConfigMapKeySelector,
	namespace string) (string, error) {
	configMap := &v1.ConfigMap{}
	nsn := types.NamespacedName{Name: selector.Name, Namespace: namespace}
	optional := selector.Optional
	err := client.Get(context.TODO(), nsn, configMap)
	if err != nil {
		if !k8sErrors.IsNotFound(err) {
			log.Errorf("Error retrieving ConfigMap %s: %v", nsn.Name, err)
			return "", err
		}
		if optional == nil || !*optional {
			err = log.ErrorfThrottledNewErr("Could not get Configmap %s from namespace %s: %v", nsn.Name, nsn.Namespace, err)
			return "", err
		}
		log.Infof("Optional Configmap %s from namespace %s not found", nsn.Name, nsn.Namespace)
		return "", nil
	}

	// Get resource data
	fieldData, ok := configMap.Data[selector.Key]
	if !ok {
		if optional == nil || !*optional {
			err := log.ErrorfThrottledNewErr("Could not get Data field %s from Resource %s from namespace %s", selector.Key, nsn.Name, nsn.Namespace)
			return "", err
		}
		log.Infof("Optional Resource %s from namespace %s missing Data key %s", nsn.Name, nsn.Namespace, selector.Key)
	}
	return fieldData, nil
}

// getSecretOverrides takes a Secret selector and returns the YAML data and handles k8s api errors appropriately
func getSecretOverrides(log vzlog.VerrazzanoLogger, client client.Client, selector *v1.SecretKeySelector,
	namespace string) (string, error) {
	sec := &v1.Secret{}
	nsn := types.NamespacedName{Name: selector.Name, Namespace: namespace}
	optional := selector.Optional
	err := client.Get(context.TODO(), nsn, sec)
	if err != nil {
		if !k8sErrors.IsNotFound(err) {
			log.Errorf("Error retrieving Secret %s: %v", nsn.Name, err)
			return "", err
		}
		if optional == nil || !*optional {
			err = log.ErrorfThrottledNewErr("Could not get Secret %s from namespace %s: %v", nsn.Name, nsn.Namespace, err)
			return "", err
		}
		log.Infof("Optional Secret %s from namespace %s not found", nsn.Name, nsn.Namespace)
		return "", nil
	}

	dataStrings := map[string]string{}
	for key, val := range sec.Data {
		dataStrings[key] = string(val)
	}

	// Get resource data
	fieldData, ok := dataStrings[selector.Key]
	if !ok {
		if optional == nil || !*optional {
			err := log.ErrorfThrottledNewErr("Could not get Data field %s from Resource %s from namespace %s", selector.Key, nsn.Name, nsn.Namespace)
			return "", err
		}
		log.Infof("Optional Resource %s from namespace %s missing Data key %s", nsn.Name, nsn.Namespace, selector.Key)
	}
	return fieldData, nil
}
