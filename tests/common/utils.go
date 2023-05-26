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
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
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
	Steps:    10,
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

func WaitForModuleToBeReady(c *v1alpha1.PlatformV1alpha1Client, namespace string, name string) (*api.Module, error) {
	var module *api.Module
	var err error
	err = Retry(DefaultRetry, vzlog.DefaultLogger(), true, func() (bool, error) {
		module, err = c.Modules(namespace).Get(context.TODO(), name, v1.GetOptions{})
		if err != nil {
			return true, err
		}

		return module.Status.State == api.ModuleStateReady, nil
	})

	return module, err
}
