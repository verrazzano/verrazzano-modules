// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package watcher

import (
	"fmt"
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
	reqs         []reconcile.Request
	t            *testing.T
	numResources int
	predicate    bool
}

// TestWatch tests that a watch can be created and that predicate return true
// GIVEN a WatchContext
// WHEN Watch is called
// THEN ensure that the resource is watched and the predicates work
func TestWatch(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name         string
		numResources int
		predicate    bool
	}{
		{
			name:         "test1",
			numResources: 0,
			predicate:    true,
		},
		{
			name:         "test2",
			numResources: 0,
			predicate:    false,
		},
		{
			name:         "test3",
			numResources: 1,
			predicate:    true,
		},
		{
			name:         "test4",
			numResources: 1,
			predicate:    false,
		},
		{
			name:         "test5",
			numResources: 2,
			predicate:    true,
		},
		{
			name:         "test6",
			numResources: 2,
			predicate:    false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := watchController{
				t:            t,
				numResources: test.numResources,
				predicate:    test.predicate,
			}
			w := &WatchContext{
				Controller:                 c,
				Log:                        vzlog.DefaultLogger(),
				ResourceKind:               source.Kind{Type: &moduleapi.Module{}},
				ShouldReconcile:            c.shouldReconcile,
				FuncGetControllerResources: c.getControllerResources,
			}
			err := w.Watch()
			asserts.NoError(err)
		})
	}
}

func (w watchController) Watch(src source.Source, eventhandler handler.EventHandler, predicates ...predicate.Predicate) error {
	asserts := assert.New(w.t)
	cr := newModuleLifecycleCR("test", "test", "")
	cr2 := newModuleLifecycleCR("test2", "test2", "")

	for _, p := range predicates {
		asserts.Equal(w.predicate, p.Create(event.CreateEvent{Object: cr}))
		asserts.Equal(w.predicate, p.Delete(event.DeleteEvent{Object: cr}))
		asserts.Equal(w.predicate, p.Update(event.UpdateEvent{ObjectOld: cr, ObjectNew: cr2}))
		asserts.False(p.Update(event.UpdateEvent{ObjectOld: cr, ObjectNew: cr}))
	}

	// Call event handler directly to get requests
	q := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	eventhandler.Create(event.CreateEvent{Object: cr}, q)
	asserts.Equal(w.numResources, q.Len())
	return nil
}

func (w watchController) getControllerResources() []types.NamespacedName {
	nsList := []types.NamespacedName{}
	for i := 1; i <= w.numResources; i++ {
		nsList = append(nsList, types.NamespacedName{
			Namespace: "namespace",
			Name:      fmt.Sprintf("name-%v", i),
		})
	}
	return nsList
}

func (w watchController) shouldReconcile(object client.Object, event spi.WatchEvent) bool {
	return w.predicate
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
