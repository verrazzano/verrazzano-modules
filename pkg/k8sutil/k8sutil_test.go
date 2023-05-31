package k8sutil

import (
	"github.com/stretchr/testify/assert"
	"github.com/verrazzano/verrazzano-modules/pkg/k8sutil/fake"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	fake2 "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestGetKubeConfigLocation(t *testing.T) {
	os.Setenv(EnvVarTestKubeConfig, "")
	os.Setenv(EnvVarKubeConfig, "")
	os.Setenv("HOME", "")

	// When environment variables are NOT set
	_, err := GetKubeConfigLocation()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "unable to find kubeconfig")

	// When TEST_KUBECONFIG is set
	expectedTestKubeConfig := "dummy-kubeconfig"
	os.Setenv(EnvVarTestKubeConfig, expectedTestKubeConfig)
	location, err := GetKubeConfigLocation()
	assert.Nil(t, err)
	assert.Equal(t, location, expectedTestKubeConfig)

	// When KUBECONFIG is set
	expectedKubeConfig := "dummy-kubeconfig"
	os.Setenv(EnvVarKubeConfig, expectedKubeConfig)
	location, err = GetKubeConfigLocation()
	assert.Nil(t, err)
	assert.Equal(t, location, expectedKubeConfig)
}

func TestGetKubeConfigGivenPath(t *testing.T) {
	kubeconfigPath := "dummy-kubeconfig"
	config, err := GetKubeConfigGivenPath(kubeconfigPath)
	assert.Nil(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, int(config.QPS), APIServerQPS)
	assert.Equal(t, config.Burst, APIServerBurst)

	fake.PodExecResult(nil)
}

func TestGetGoClient(t *testing.T) {
	config, err := buildRESTConfig(vzlog.DefaultLogger())
	assert.Nil(t, err)
	assert.NotNil(t, config)

	result, err := GetGoClient(vzlog.DefaultLogger())
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestGetURLForIngress(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = networkingv1.AddToScheme(scheme)

	t.Run("Test GetIngressURL success", func(t *testing.T) {
		ingress := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-ingress",
				Namespace: "testns",
			},
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{
					{
						Host: "test.com",
					},
				},
			},
		}

		cl := fake2.NewFakeClientWithScheme(scheme, ingress)
		url, err := GetURLForIngress(cl, "test-ingress", "testns", "https")
		assert.Nil(t, err)
		assert.NotNil(t, url)
	})

	t.Run("Test GetIngressURL fail", func(t *testing.T) {
		cl := fake2.NewFakeClientWithScheme(scheme)
		url, err := GetURLForIngress(cl, "dummy-ingress", "testns", "https")
		assert.NotNil(t, err)
		assert.Equal(t, len(url), 0)
	})

	t.Run("Test GetIngressURL no rules", func(t *testing.T) {
		ingress := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-ingress",
				Namespace: "testns",
			},
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{
					{},
				},
			},
		}
		cl := fake2.NewFakeClientWithScheme(scheme, ingress)
		url, err := GetURLForIngress(cl, "test-ingress", "testns", "https")
		assert.Nil(t, err)
		assert.Equal(t, url, "https://")
	})
}
