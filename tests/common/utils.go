// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package common

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	api "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/clientset/versioned/typed/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/pkg/k8sutil"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	corev1 "k8s.io/api/core/v1"
	kerrs "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func LoadTestFile(filePath string) ([]byte, error) {
	testDataDir := os.Getenv("TESTDATA_DIR")
	if testDataDir == "" {
		return nil, fmt.Errorf("TESTDATA_DIR not defined")
	}
	fileName := filepath.Join(testDataDir, filePath)
	if _, err := os.Stat(fileName); err != nil {
		return nil, fmt.Errorf("unable to read test file %s, err: %v", fileName, err)
	}
	return os.ReadFile(fileName)
}

func UnmarshalTestFile(filePath string, element interface{}) error {
	data, err := LoadTestFile(filePath)
	if err != nil {
		return fmt.Errorf("unable to load test file data for %s, err: %v", filePath, err)
	}

	return yaml.Unmarshal(data, element)
}

var DefaultRetry = wait.Backoff{
	Steps:    5,
	Duration: 1 * time.Second,
	Factor:   2.0,
	Jitter:   0.1,
}

// Retry executes the provided function repeatedly, retrying until the function
// returns done = true, or exceeds the given timeout.
// errors will be logged, but will not trigger retry to stop unless retryOnError is false
func Retry(backoff wait.Backoff, log vzlog.VerrazzanoLogger, retryOnError bool, fn wait.ConditionFunc) error {
	var lastErr error
	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		done, err := fn()
		lastErr = err
		if err != nil && retryOnError {
			log.Infof("Retrying after error: %v", err)
			return done, nil
		}
		return done, err
	})
	if err == wait.ErrWaitTimeout {
		if lastErr != nil {
			err = lastErr
		}
	}
	return err
}

func WaitForModuleToBeReady(c *v1alpha1.PlatformV1alpha1Client, module *api.Module) (*api.Module, error) {
	var deployedModule *api.Module
	var err, retryError error
	retryError = Retry(DefaultRetry, vzlog.DefaultLogger(), true, func() (bool, error) {
		deployedModule, err = c.Modules(module.GetNamespace()).Get(context.TODO(), module.GetName(), v1.GetOptions{})
		if err != nil {
			return false, err
		}
		return deployedModule.Status.State == api.ModuleStateReady, nil
	})

	return deployedModule, retryError
}

func WaitForModuleToBeDeleted(c *v1alpha1.PlatformV1alpha1Client, namespace string, name string) error {
	err := Retry(DefaultRetry, vzlog.DefaultLogger(), true, func() (bool, error) {
		_, err := c.Modules(namespace).Get(context.TODO(), name, v1.GetOptions{})
		if err != nil {
			if kerrs.IsNotFound(err) {
				return true, nil
			}

			return false, err
		}

		return false, nil
	})

	return err
}

func WaitForModuleToBeUpgraded(c *v1alpha1.PlatformV1alpha1Client, module *api.Module) (*api.Module, error) {
	var deployedModule *api.Module
	var err, retryError error
	retryError = Retry(DefaultRetry, vzlog.DefaultLogger(), true, func() (bool, error) {
		deployedModule, err = c.Modules(module.GetNamespace()).Get(context.TODO(), module.GetName(), v1.GetOptions{})
		if err != nil {
			return false, err
		}
		return deployedModule.Status.State == api.ModuleStateReady && deployedModule.Status.Version == module.Spec.Version, nil
	})

	return deployedModule, retryError
}

func WaitForNamespaceCreated(namespace string) error {
	c, err := k8sutil.GetCoreV1Client()
	if err != nil {
		return err
	}

	err = Retry(DefaultRetry, vzlog.DefaultLogger(), true, func() (bool, error) {
		_, err := c.Namespaces().Get(context.TODO(), namespace, v1.GetOptions{})
		if err != nil {
			if kerrs.IsNotFound(err) {
				_, err = c.Namespaces().Create(context.TODO(), &corev1.Namespace{ObjectMeta: v1.ObjectMeta{Name: namespace}}, v1.CreateOptions{})
				return true, err
			}

			return false, err
		}

		return true, nil
	})

	return err
}

func WaitForNamespaceDeleted(namespace string) error {
	c, err := k8sutil.GetCoreV1Client()
	if err != nil {
		return err
	}

	err = Retry(DefaultRetry, vzlog.DefaultLogger(), true, func() (bool, error) {
		_, err := c.Namespaces().Get(context.TODO(), namespace, v1.GetOptions{})
		if err != nil {
			if kerrs.IsNotFound(err) {
				return true, nil
			}

			return false, err
		}

		return false, nil
	})

	return err
}
