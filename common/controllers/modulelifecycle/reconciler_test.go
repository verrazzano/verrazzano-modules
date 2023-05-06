// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package modulelifecycle

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/verrazzano/verrazzano-modules/common/actionspi"
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/spi"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano/pkg/log/vzlog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	fakes "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

const namespace = "testns"
const name = "test"

func TestReconcile(t *testing.T) {
	asserts := assert.New(t)

	rctx := spi.ReconcileContext{
		Log:       vzlog.DefaultLogger(),
		ClientCtx: context.TODO(),
	}

	cr := newModuleLifecycleCR
	if u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(cr); err != nil {
		return ctrl.Result{}, err
	}
	scheme := initScheme()
	clientBuilder := fakes.NewClientBuilder().WithScheme(scheme)

	r := Reconciler{
		Client:   nil,
		Scheme:   initScheme(),
		handlers: actionspi.ActionHandlers{},
	}

}

func newModuleLifecycleCR(namespace string, name string, className string) *moduleapi.ModuleLifecycle {
	m := &moduleapi.ModuleLifecycle{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: moduleapi.ModuleLifecycleSpec{
			LifecycleClassName: moduleapi.LifecycleClassType(className),
		},
	}
	return m
}

func initScheme() *runtime.Scheme {
	// Create a scheme then add each GKV group to the scheme
	scheme := runtime.NewScheme()

	utilruntime.Must(moduleapi.AddToScheme(scheme))
	return scheme
}
