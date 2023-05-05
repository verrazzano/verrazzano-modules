// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package basecontroller

import (
	crbuilder "sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type FakeBuilder struct{}

func (blder *FakeBuilder) For(object client.Object, opts ...crbuilder.ForOption) *FakeBuilder {
	return blder
}

func (blder *FakeBuilder) WithEventFilter(p predicate.Predicate) *FakeBuilder {
	return blder
}

func (blder *FakeBuilder) Build(r Reconciler) (controller.Controller, error) {
	return fakeController{}, nil
}
