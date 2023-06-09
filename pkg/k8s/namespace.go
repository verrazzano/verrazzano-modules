// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package k8s

import (
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func GetNamespacedName(meta v1.ObjectMeta) types.NamespacedName {
	return types.NamespacedName{
		Namespace: meta.Namespace,
		Name:      meta.Name,
	}
}

func GetNamespacedNameString(meta v1.ObjectMeta) string {
	return fmt.Sprintf("%s/%s", meta.Namespace, meta.Name)
}
