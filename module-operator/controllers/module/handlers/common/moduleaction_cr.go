// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package common

import (
	"context"
	"fmt"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/handlerspi"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// GetModuleAction gets a ModuleAction CR
func (h BaseHandler) GetModuleAction(ctx handlerspi.HandlerContext) (*moduleapi.ModuleAction, error) {
	moduleAction := moduleapi.ModuleAction{}
	nsn := types.NamespacedName{
		Name:      h.ModuleActionName,
		Namespace: h.ModuleCR.Namespace,
	}

	if err := ctx.Client.Get(context.TODO(), nsn, &moduleAction); err != nil {
		ctx.Log.Progressf("Retrying get for ModuleAction %v: %v", nsn, err)
		return nil, err
	}
	return &moduleAction, nil
}

// DeleteModuleAction deletes a ModuleAction CR
func (h BaseHandler) DeleteModuleAction(ctx handlerspi.HandlerContext) error {
	moduleAction := moduleapi.ModuleAction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      h.ModuleActionName,
			Namespace: h.ModuleCR.Namespace,
		},
	}

	if err := ctx.Client.Delete(context.TODO(), &moduleAction); err != nil {
		ctx.Log.ErrorfThrottled("Failed trying to delete ModuleAction %s/%s: %v", moduleAction.Namespace, moduleAction.Name, err)
		return err
	}
	return nil
}

// BuildModuleActionCRName builds a ModuleAction CR name
func BuildModuleActionCRName(moduleCRName string, lifecycleClassName moduleapi.ModuleClassType, action moduleapi.ModuleActionType) string {
	return fmt.Sprintf("%s-%s-%s", moduleCRName, lifecycleClassName, action)
}
