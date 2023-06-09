// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeModules implements ModuleInterface
type FakeModules struct {
	Fake *FakePlatformV1alpha1
	ns   string
}

var modulesResource = schema.GroupVersionResource{Group: "platform", Version: "v1alpha1", Resource: "modules"}

var modulesKind = schema.GroupVersionKind{Group: "platform", Version: "v1alpha1", Kind: "Module"}

// Get takes name of the module, and returns the corresponding module object, and an error if there is any.
func (c *FakeModules) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Module, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(modulesResource, c.ns, name), &v1alpha1.Module{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Module), err
}

// List takes label and field selectors, and returns the list of Modules that match those selectors.
func (c *FakeModules) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ModuleList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(modulesResource, modulesKind, c.ns, opts), &v1alpha1.ModuleList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ModuleList{ListMeta: obj.(*v1alpha1.ModuleList).ListMeta}
	for _, item := range obj.(*v1alpha1.ModuleList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested modules.
func (c *FakeModules) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(modulesResource, c.ns, opts))

}

// Create takes the representation of a module and creates it.  Returns the server's representation of the module, and an error, if there is any.
func (c *FakeModules) Create(ctx context.Context, module *v1alpha1.Module, opts v1.CreateOptions) (result *v1alpha1.Module, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(modulesResource, c.ns, module), &v1alpha1.Module{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Module), err
}

// Update takes the representation of a module and updates it. Returns the server's representation of the module, and an error, if there is any.
func (c *FakeModules) Update(ctx context.Context, module *v1alpha1.Module, opts v1.UpdateOptions) (result *v1alpha1.Module, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(modulesResource, c.ns, module), &v1alpha1.Module{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Module), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeModules) UpdateStatus(ctx context.Context, module *v1alpha1.Module, opts v1.UpdateOptions) (*v1alpha1.Module, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(modulesResource, "status", c.ns, module), &v1alpha1.Module{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Module), err
}

// Delete takes name of the module and deletes it. Returns an error if one occurs.
func (c *FakeModules) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(modulesResource, c.ns, name, opts), &v1alpha1.Module{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeModules) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(modulesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.ModuleList{})
	return err
}

// Patch applies the patch and returns the patched module.
func (c *FakeModules) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Module, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(modulesResource, c.ns, name, pt, data, subresources...), &v1alpha1.Module{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Module), err
}
