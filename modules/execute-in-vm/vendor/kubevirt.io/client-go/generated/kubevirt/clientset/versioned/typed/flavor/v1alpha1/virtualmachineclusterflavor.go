/*
Copyright 2022 The KubeVirt Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"

	v1alpha1 "kubevirt.io/api/flavor/v1alpha1"
	scheme "kubevirt.io/client-go/generated/kubevirt/clientset/versioned/scheme"
)

// VirtualMachineClusterFlavorsGetter has a method to return a VirtualMachineClusterFlavorInterface.
// A group's client should implement this interface.
type VirtualMachineClusterFlavorsGetter interface {
	VirtualMachineClusterFlavors() VirtualMachineClusterFlavorInterface
}

// VirtualMachineClusterFlavorInterface has methods to work with VirtualMachineClusterFlavor resources.
type VirtualMachineClusterFlavorInterface interface {
	Create(ctx context.Context, virtualMachineClusterFlavor *v1alpha1.VirtualMachineClusterFlavor, opts v1.CreateOptions) (*v1alpha1.VirtualMachineClusterFlavor, error)
	Update(ctx context.Context, virtualMachineClusterFlavor *v1alpha1.VirtualMachineClusterFlavor, opts v1.UpdateOptions) (*v1alpha1.VirtualMachineClusterFlavor, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.VirtualMachineClusterFlavor, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.VirtualMachineClusterFlavorList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.VirtualMachineClusterFlavor, err error)
	VirtualMachineClusterFlavorExpansion
}

// virtualMachineClusterFlavors implements VirtualMachineClusterFlavorInterface
type virtualMachineClusterFlavors struct {
	client rest.Interface
}

// newVirtualMachineClusterFlavors returns a VirtualMachineClusterFlavors
func newVirtualMachineClusterFlavors(c *FlavorV1alpha1Client) *virtualMachineClusterFlavors {
	return &virtualMachineClusterFlavors{
		client: c.RESTClient(),
	}
}

// Get takes name of the virtualMachineClusterFlavor, and returns the corresponding virtualMachineClusterFlavor object, and an error if there is any.
func (c *virtualMachineClusterFlavors) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.VirtualMachineClusterFlavor, err error) {
	result = &v1alpha1.VirtualMachineClusterFlavor{}
	err = c.client.Get().
		Resource("virtualmachineclusterflavors").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of VirtualMachineClusterFlavors that match those selectors.
func (c *virtualMachineClusterFlavors) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.VirtualMachineClusterFlavorList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.VirtualMachineClusterFlavorList{}
	err = c.client.Get().
		Resource("virtualmachineclusterflavors").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested virtualMachineClusterFlavors.
func (c *virtualMachineClusterFlavors) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("virtualmachineclusterflavors").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a virtualMachineClusterFlavor and creates it.  Returns the server's representation of the virtualMachineClusterFlavor, and an error, if there is any.
func (c *virtualMachineClusterFlavors) Create(ctx context.Context, virtualMachineClusterFlavor *v1alpha1.VirtualMachineClusterFlavor, opts v1.CreateOptions) (result *v1alpha1.VirtualMachineClusterFlavor, err error) {
	result = &v1alpha1.VirtualMachineClusterFlavor{}
	err = c.client.Post().
		Resource("virtualmachineclusterflavors").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(virtualMachineClusterFlavor).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a virtualMachineClusterFlavor and updates it. Returns the server's representation of the virtualMachineClusterFlavor, and an error, if there is any.
func (c *virtualMachineClusterFlavors) Update(ctx context.Context, virtualMachineClusterFlavor *v1alpha1.VirtualMachineClusterFlavor, opts v1.UpdateOptions) (result *v1alpha1.VirtualMachineClusterFlavor, err error) {
	result = &v1alpha1.VirtualMachineClusterFlavor{}
	err = c.client.Put().
		Resource("virtualmachineclusterflavors").
		Name(virtualMachineClusterFlavor.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(virtualMachineClusterFlavor).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the virtualMachineClusterFlavor and deletes it. Returns an error if one occurs.
func (c *virtualMachineClusterFlavors) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("virtualmachineclusterflavors").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *virtualMachineClusterFlavors) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("virtualmachineclusterflavors").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched virtualMachineClusterFlavor.
func (c *virtualMachineClusterFlavors) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.VirtualMachineClusterFlavor, err error) {
	result = &v1alpha1.VirtualMachineClusterFlavor{}
	err = c.client.Patch(pt).
		Resource("virtualmachineclusterflavors").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}