// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package basecontroller

import (
	"context"
	"github.com/stretchr/testify/assert"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakes "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"testing"
)

type ManagerReconcilerImpl struct {
	className string
}

// TestCreatePredicateFilter tests the creation of the predicate filter
// GIVEN a base controller
// WHEN CreatePredicateFilter is called
// THEN the filter should be properly initialized
func TestCreatePredicateFilter(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name                string
		className           string
		crClassName         string
		expectedHandlerBool bool
	}{
		{
			name:                "test1",
			className:           "",
			crClassName:         "",
			expectedHandlerBool: true,
		},
		{
			name:                "test2",
			className:           "myclass",
			crClassName:         "myclass",
			expectedHandlerBool: true,
		},
		{
			name:                "test2",
			className:           "myclass",
			crClassName:         "",
			expectedHandlerBool: false,
		},
		{
			name:                "test2",
			className:           "",
			crClassName:         "nyclass",
			expectedHandlerBool: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reconciler := ManagerReconcilerImpl{className: test.className}
			config := ControllerConfig{
				EventFilter: &reconciler,
			}
			cr := newModuleCR(namespace, name)
			cr.Spec.ModuleName = test.crClassName
			clientBuilder := fakes.NewClientBuilder()
			c := clientBuilder.WithScheme(newScheme()).WithObjects(cr).Build()
			r := newReconciler(c, config)
			f := r.createPredicateFilter(r.layeredControllerConfig.EventFilter)

			asserts.NotNil(f.Create)
			asserts.NotNil(f.Delete)
			asserts.NotNil(f.Update)
			asserts.NotNil(f.Generic)

			asserts.Equal(test.expectedHandlerBool, f.Create(event.CreateEvent{Object: cr}))
			asserts.Equal(test.expectedHandlerBool, f.Delete(event.DeleteEvent{Object: cr}))
			asserts.Equal(test.expectedHandlerBool, f.Update(event.UpdateEvent{ObjectOld: cr, ObjectNew: cr}))
			asserts.Equal(test.expectedHandlerBool, f.Generic(event.GenericEvent{Object: cr}))
		})
	}
}

func (r *ManagerReconcilerImpl) HandlePredicateEvent(cli client.Client, object client.Object) bool {
	mlc := moduleapi.Module{}
	objectkey := client.ObjectKeyFromObject(object)
	if err := cli.Get(context.TODO(), objectkey, &mlc); err != nil {
		return false
	}
	return mlc.Spec.ModuleName == string(r.className)
}
