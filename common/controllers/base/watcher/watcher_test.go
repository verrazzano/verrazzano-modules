// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package watcher

import (
	"github.com/stretchr/testify/assert"
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/fake"
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/spi"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano/pkg/log/vzlog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"testing"
)

type watchController struct {
	fake.FakeController
	reqs []reconcile.Request
	t    *testing.T
}

// TestWatch tests that a watch can be created and that predicate return true
// GIVEN a WatchContext
// WHEN Watch is called
// THEN ensure that the resource is watched and the predicates work
func TestWatch(t *testing.T) {
	asserts := assert.New(t)

	c := watchController{t: t}
	w := &WatchContext{
		Controller:                 c,
		Log:                        vzlog.DefaultLogger(),
		ResourceKind:               source.Kind{Type: &moduleapi.Module{}},
		ShouldReconcile:            shouldReconcile,
		FuncGetControllerResources: getControllerResources,
	}
	err := w.Watch()
	asserts.NoError(err)
}

func (w watchController) Watch(src source.Source, eventhandler handler.EventHandler, predicates ...predicate.Predicate) error {
	asserts := assert.New(w.t)
	cr := newModuleLifecycleCR("test", "test", "")
	cr2 := newModuleLifecycleCR("test2", "test2", "")

	for _, p := range predicates {
		asserts.True(p.Create(event.CreateEvent{Object: cr}))
		asserts.True(p.Delete(event.DeleteEvent{Object: cr}))
		asserts.True(p.Generic(event.GenericEvent{Object: cr}))
		asserts.False(p.Update(event.UpdateEvent{ObjectOld: cr, ObjectNew: cr}))
		asserts.True(p.Update(event.UpdateEvent{ObjectOld: cr, ObjectNew: cr2}))
	}

	// Call event handler directly to get requests
	q := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	eventhandler.Create(event.CreateEvent{Object: cr}, q)
	asserts.Equal(1, q.Len())
	return nil
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

func getControllerResources() []types.NamespacedName {
	return []types.NamespacedName{{
		Namespace: "namespace",
		Name:      "name",
	}}
}

func shouldReconcile(object client.Object, event spi.WatchEvent) bool {
	return true
}
