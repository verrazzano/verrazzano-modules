// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package common

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/verrazzano/verrazzano-modules/module-operator/clientset/versioned/typed/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/pkg/k8sutil"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// TestContext contains the reference to testing.T for current test state,
// k8s clients and gomega instance specific to running test.
type TestContext struct {
	T            *testing.T
	GomegaWithT  *gomega.WithT
	c            *v1alpha1.PlatformV1alpha1Client
	corev1client v1.CoreV1Interface
	client       *kubernetes.Clientset
}

// NewTestContext creates a new TestContext.
func NewTestContext(t *testing.T) *TestContext {
	return &TestContext{
		T:           t,
		GomegaWithT: gomega.NewWithT(t),
	}
}

// ModuleClient returns the k8s client for Modules.
func (ctx *TestContext) ModuleClient() *v1alpha1.PlatformV1alpha1Client {
	if ctx.c == nil {
		config, err := k8sutil.GetKubeConfig()
		ctx.GomegaWithT.Expect(err).NotTo(gomega.HaveOccurred())
		ctx.c, err = v1alpha1.NewForConfig(config)
		ctx.GomegaWithT.Expect(err).NotTo(gomega.HaveOccurred())
	}

	return ctx.c
}

// CoreV1Client returns the k8s client for corev1 resources.
func (ctx *TestContext) CoreV1Client() v1.CoreV1Interface {
	if ctx.corev1client == nil {
		var err error
		ctx.corev1client, err = k8sutil.GetCoreV1Client(vzlog.DefaultLogger())
		ctx.GomegaWithT.Expect(err).NotTo(gomega.HaveOccurred())
	}

	return ctx.corev1client
}

// Client returns the generic k8s client.
func (ctx *TestContext) Client() *kubernetes.Clientset {
	if ctx.client == nil {
		var err error
		ctx.client, err = k8sutil.GetKubernetesClientset()
		ctx.GomegaWithT.Expect(err).NotTo(gomega.HaveOccurred())
	}

	return ctx.client
}
