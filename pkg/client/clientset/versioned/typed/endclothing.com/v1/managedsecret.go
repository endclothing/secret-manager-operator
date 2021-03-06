// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"time"

	v1 "github.com/endclothing/secret-manager-operator/pkg/apis/endclothing.com/v1"
	scheme "github.com/endclothing/secret-manager-operator/pkg/client/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ManagedSecretsGetter has a method to return a ManagedSecretInterface.
// A group's client should implement this interface.
type ManagedSecretsGetter interface {
	ManagedSecrets(namespace string) ManagedSecretInterface
}

// ManagedSecretInterface has methods to work with ManagedSecret resources.
type ManagedSecretInterface interface {
	Create(*v1.ManagedSecret) (*v1.ManagedSecret, error)
	Update(*v1.ManagedSecret) (*v1.ManagedSecret, error)
	UpdateStatus(*v1.ManagedSecret) (*v1.ManagedSecret, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error
	Get(name string, options metav1.GetOptions) (*v1.ManagedSecret, error)
	List(opts metav1.ListOptions) (*v1.ManagedSecretList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.ManagedSecret, err error)
	ManagedSecretExpansion
}

// managedSecrets implements ManagedSecretInterface
type managedSecrets struct {
	client rest.Interface
	ns     string
}

// newManagedSecrets returns a ManagedSecrets
func newManagedSecrets(c *ComV1Client, namespace string) *managedSecrets {
	return &managedSecrets{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the managedSecret, and returns the corresponding managedSecret object, and an error if there is any.
func (c *managedSecrets) Get(name string, options metav1.GetOptions) (result *v1.ManagedSecret, err error) {
	result = &v1.ManagedSecret{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("managedsecrets").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ManagedSecrets that match those selectors.
func (c *managedSecrets) List(opts metav1.ListOptions) (result *v1.ManagedSecretList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.ManagedSecretList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("managedsecrets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested managedSecrets.
func (c *managedSecrets) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("managedsecrets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a managedSecret and creates it.  Returns the server's representation of the managedSecret, and an error, if there is any.
func (c *managedSecrets) Create(managedSecret *v1.ManagedSecret) (result *v1.ManagedSecret, err error) {
	result = &v1.ManagedSecret{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("managedsecrets").
		Body(managedSecret).
		Do().
		Into(result)
	return
}

// Update takes the representation of a managedSecret and updates it. Returns the server's representation of the managedSecret, and an error, if there is any.
func (c *managedSecrets) Update(managedSecret *v1.ManagedSecret) (result *v1.ManagedSecret, err error) {
	result = &v1.ManagedSecret{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("managedsecrets").
		Name(managedSecret.Name).
		Body(managedSecret).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *managedSecrets) UpdateStatus(managedSecret *v1.ManagedSecret) (result *v1.ManagedSecret, err error) {
	result = &v1.ManagedSecret{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("managedsecrets").
		Name(managedSecret.Name).
		SubResource("status").
		Body(managedSecret).
		Do().
		Into(result)
	return
}

// Delete takes name of the managedSecret and deletes it. Returns an error if one occurs.
func (c *managedSecrets) Delete(name string, options *metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("managedsecrets").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *managedSecrets) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("managedsecrets").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched managedSecret.
func (c *managedSecrets) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.ManagedSecret, err error) {
	result = &v1.ManagedSecret{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("managedsecrets").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
