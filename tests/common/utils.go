// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package common

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/util/yaml"
)

// LoadTestFile reads a testdata file.
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

// UnmarshalTestFile unmarshalls a testdata file to a go object.
func UnmarshalTestFile(filePath string, element interface{}) error {
	data, err := LoadTestFile(filePath)
	if err != nil {
		return fmt.Errorf("unable to load test file data for %s, err: %v", filePath, err)
	}

	return yaml.Unmarshal(data, element)
}
