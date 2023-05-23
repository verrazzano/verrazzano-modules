// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package basecontroller

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/base/controllerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	controllerruntime "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"testing"
)

type watchController struct {
	FakeController
	t            *testing.T
	numResources int
	predicate    bool
}

type FakeController struct{}

var _ controllerruntime.Controller = &FakeController{}

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
	cr := newModuleCR("test", "test")
	cr2 := newModuleCR("test2", "test2")

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

func (w watchController) shouldReconcile(object client.Object, event controllerspi.WatchEvent) bool {
	return w.predicate
}

func (f FakeController) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	return ctrl.Result{}, nil
}

func (f FakeController) Watch(src source.Source, eventhandler handler.EventHandler, predicates ...predicate.Predicate) error {
	return nil
}

func (f FakeController) Start(ctx context.Context) error {
	return nil
}

func (f FakeController) GetLogger() logr.Logger {
	return logr.Logger{}
}
