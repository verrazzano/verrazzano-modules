// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package common

import (
	"fmt"
	moduleplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
)

func DeriveModuleLifeCycleName(moduleCRName string, lifecycleClassName moduleplatform.LifecycleClassType, action moduleplatform.ActionType) string {
	return fmt.Sprintf("%s-%s-%s", moduleCRName, lifecycleClassName, action)
}
