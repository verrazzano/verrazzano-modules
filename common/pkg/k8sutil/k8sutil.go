// Copyright (c) 2021, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package k8sutil

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	k8sversionutil "k8s.io/apimachinery/pkg/util/version"
	"os"
	"path/filepath"
	"strings"

	"github.com/verrazzano/verrazzano-modules/common/pkg/vzlog"

	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	kerrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/homedir"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EnvVarKubeConfig Name of Environment Variable for KUBECONFIG
const EnvVarKubeConfig = "KUBECONFIG"

// EnvVarTestKubeConfig Name of Environment Variable for test KUBECONFIG
const EnvVarTestKubeConfig = "TEST_KUBECONFIG"

const APIServerBurst = 150
const APIServerQPS = 100

type ClientConfigFunc func() (*rest.Config, kubernetes.Interface, error)

var ClientConfig ClientConfigFunc = func() (*rest.Config, kubernetes.Interface, error) {
	cfg, err := GetConfigFromController()
	if err != nil {
		return nil, nil, err
	}
	c, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, nil, err
	}
	return cfg, c, nil
}

// fakeClient is for unit testing
var fakeClient kubernetes.Interface

// SetFakeClient for unit tests
func SetFakeClient(client kubernetes.Interface) {
	fakeClient = client
}

// ClearFakeClient for unit tests
func ClearFakeClient() {
	fakeClient = nil
}

// NewControllerRuntimeClient Create a new controller runtime client
func NewControllerRuntimeClient(opts client.Options) (client.Client, error) {
	config, err := GetConfigFromController()
	if err != nil {
		return nil, err
	}

	return client.New(config, opts)
}

// GetConfigFromController get the config from the Controller Runtime and set the default QPS and burst.
func GetConfigFromController() (*rest.Config, error) {
	cfg, err := controllerruntime.GetConfig()
	if err != nil {
		return nil, err
	}
	setConfigQPSBurst(cfg)
	return cfg, nil
}

// GetConfigOrDieFromController get the config from the Controller Runtime and set the default QPS and burst.
func GetConfigOrDieFromController() *rest.Config {
	cfg := controllerruntime.GetConfigOrDie()
	setConfigQPSBurst(cfg)
	return cfg
}

// GetKubeConfigLocation Helper function to obtain the default kubeConfig location
func GetKubeConfigLocation() (string, error) {
	if testKubeConfig := os.Getenv(EnvVarTestKubeConfig); len(testKubeConfig) > 0 {
		return testKubeConfig, nil
	}

	if kubeConfig := os.Getenv(EnvVarKubeConfig); len(kubeConfig) > 0 {
		return kubeConfig, nil
	}

	if home := homedir.HomeDir(); home != "" {
		return filepath.Join(home, ".kube", "config"), nil
	}

	return "", errors.New("unable to find kubeconfig")

}

// GetKubeConfigGivenPath GetKubeConfig will get the kubeconfig from the given kubeconfigPath
func GetKubeConfigGivenPath(kubeconfigPath string) (*rest.Config, error) {
	return BuildKubeConfig(kubeconfigPath)
}

func BuildKubeConfig(kubeconfig string) (*rest.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	setConfigQPSBurst(config)
	return config, nil
}

// GetKubeConfig Returns kubeconfig from KUBECONFIG env var if set
// Else from default location ~/.kube/config
func GetKubeConfig() (*rest.Config, error) {
	var config *rest.Config
	kubeConfigLoc, err := GetKubeConfigLocation()
	if err != nil {
		return nil, err
	}
	config, err = clientcmd.BuildConfigFromFlags("", kubeConfigLoc)
	if err != nil {
		return nil, err
	}
	setConfigQPSBurst(config)
	return config, nil
}

// GetKubeConfigGivenPathAndContext returns a rest.Config given a kubeConfig and kubeContext.
func GetKubeConfigGivenPathAndContext(kubeConfigPath string, kubeContext string) (*rest.Config, error) {
	// If no values passed, call default GetKubeConfig
	if len(kubeConfigPath) == 0 && len(kubeContext) == 0 {
		return GetKubeConfig()
	}

	// Default the value of kubeConfigLoc?
	var err error
	if len(kubeConfigPath) == 0 {
		kubeConfigPath, err = GetKubeConfigLocation()
		if err != nil {
			return nil, err
		}
	}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeConfigPath},
		&clientcmd.ConfigOverrides{CurrentContext: kubeContext}).ClientConfig()
	if err != nil {
		return nil, err
	}
	setConfigQPSBurst(config)
	return config, nil
}

// GetKubernetesClientset returns the Kubernetes clientset for the cluster set in the environment
func GetKubernetesClientset() (*kubernetes.Clientset, error) {
	// use the current context in the kubeconfig
	var clientset *kubernetes.Clientset
	config, err := GetKubeConfig()
	if err != nil {
		return clientset, err
	}
	return GetKubernetesClientsetWithConfig(config)
}

// GetKubernetesClientsetOrDie returns the kubernetes clientset, panic if it cannot be created.
func GetKubernetesClientsetOrDie() *kubernetes.Clientset {
	clientset, err := GetKubernetesClientset()
	if err != nil {
		panic(err)
	}
	return clientset
}

// GetKubernetesClientsetWithConfig returns the Kubernetes clientset for the given configuration
func GetKubernetesClientsetWithConfig(config *rest.Config) (*kubernetes.Clientset, error) {
	var clientset *kubernetes.Clientset
	clientset, err := kubernetes.NewForConfig(config)
	return clientset, err
}

// GetCoreV1Func is the function to return the CoreV1Interface
var GetCoreV1Func = GetCoreV1Client

// GetCoreV1Client returns the CoreV1Interface
func GetCoreV1Client(log ...vzlog.VerrazzanoLogger) (corev1.CoreV1Interface, error) {
	goClient, err := GetGoClient(log...)
	if err != nil {
		return nil, err
	}
	return goClient.CoreV1(), nil
}

// GetAppsV1Func is the function the AppsV1Interface
var GetAppsV1Func = GetAppsV1Client

// GetAppsV1Client returns the AppsV1Interface
func GetAppsV1Client(log ...vzlog.VerrazzanoLogger) (appsv1.AppsV1Interface, error) {
	goClient, err := GetGoClient(log...)
	if err != nil {
		return nil, err
	}
	return goClient.AppsV1(), nil
}

// GetDynamicClientFunc is the function to return the Dynamic Interface
var GetDynamicClientFunc = GetDynamicClient

// GetDynamicClient returns the Dynamic Interface
func GetDynamicClient() (dynamic.Interface, error) {
	config, err := GetConfigFromController()
	if err != nil {
		return nil, err
	}
	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return dynClient, nil
}

// NewPodExecutor is to be overridden during unit tests
var NewPodExecutor = remotecommand.NewSPDYExecutor

// ExecPod runs a remote command a pod, returning the stdout and stderr of the command.
func ExecPod(client kubernetes.Interface, cfg *rest.Config, pod *v1.Pod, container string, command []string) (string, string, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	request := client.
		CoreV1().
		RESTClient().
		Post().
		Namespace(pod.Namespace).
		Resource("pods").
		Name(pod.Name).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: container,
			Command:   command,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)
	executor, err := NewPodExecutor(cfg, "POST", request.URL())
	if err != nil {
		return "", "", err
	}
	err = executor.Stream(remotecommand.StreamOptions{
		Stdout: stdout,
		Stderr: stderr,
	})
	if err != nil {
		return "", "", fmt.Errorf("error running command %s on %v/%v: %v", command, pod.Namespace, pod.Name, err)
	}

	return stdout.String(), stderr.String(), nil
}

// ExecPodNoTty runs a remote command a pod, returning the stdout and stderr of the command.
func ExecPodNoTty(client kubernetes.Interface, cfg *rest.Config, pod *v1.Pod, container string, command []string) (string, string, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	request := client.
		CoreV1().
		RESTClient().
		Post().
		Namespace(pod.Namespace).
		Resource("pods").
		Name(pod.Name).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: container,
			Command:   command,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)
	executor, err := NewPodExecutor(cfg, "POST", request.URL())
	if err != nil {
		return "", "", err
	}
	err = executor.Stream(remotecommand.StreamOptions{
		Stdout: stdout,
		Stderr: stderr,
	})
	if err != nil {
		return "", "", fmt.Errorf("error running command %s on %v/%v: %v", command, pod.Namespace, pod.Name, err)
	}

	return stdout.String(), stderr.String(), nil
}

// GetGoClient returns a go-client
func GetGoClient(log ...vzlog.VerrazzanoLogger) (kubernetes.Interface, error) {
	if fakeClient != nil {
		return fakeClient, nil
	}
	var logger vzlog.VerrazzanoLogger
	if len(log) > 0 {
		logger = log[0]
	}
	config, err := buildRESTConfig(logger)
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		if logger != nil {
			logger.Errorf("Failed to get clientset: %v", err)
		}
		return nil, err
	}

	return kubeClient, err
}

func buildRESTConfig(logger vzlog.VerrazzanoLogger) (*rest.Config, error) {
	config, err := GetConfigFromController()
	if err != nil {
		if logger != nil {
			logger.Errorf("Failed to get kubeconfig: %v", err)
		}
		return nil, err
	}
	return config, nil
}

// GetDynamicClientInCluster returns a dynamic client needed to access Unstructured data
func GetDynamicClientInCluster(kubeconfigPath string) (dynamic.Interface, error) {
	config, err := GetKubeConfigGivenPath(kubeconfigPath)
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(config)
}

// GetURLForIngress returns the url for an Ingress
func GetURLForIngress(client client.Client, name string, namespace string, scheme string) (string, error) {
	var ingress = &networkingv1.Ingress{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, ingress)
	if err != nil {
		return "", fmt.Errorf("unable to fetch ingress %s/%s, %v", name, namespace, err)
	}
	return fmt.Sprintf("%s://%s", scheme, ingress.Spec.Rules[0].Host), nil
}

// GetRunningPodForLabel returns the reference of a running pod that matches the given label
func GetRunningPodForLabel(c client.Client, label string, namespace string, log ...vzlog.VerrazzanoLogger) (*v1.Pod, error) {
	var logger vzlog.VerrazzanoLogger
	if len(log) > 0 {
		logger = log[0]
	} else {
		logger = vzlog.DefaultLogger()
	}

	pods := &v1.PodList{}
	labelPair := strings.Split(label, "=")
	err := c.List(context.Background(), pods, client.MatchingLabels{labelPair[0]: labelPair[1]})

	if err != nil {
		return nil, logger.ErrorfThrottledNewErr("Failed getting running pods for label %s in namespace %s, error: %v", label, namespace, err.Error())
	}

	if !(len(pods.Items) > 0) {
		return nil, logger.ErrorfThrottledNewErr("Invalid running pod list for label %s in namespace %s", label, namespace)
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodRunning {
			return &pod, nil
		}
	}

	return nil, logger.ErrorfThrottledNewErr("No running pod for label %s in namespace %s", label, namespace)
}

// ErrorIfDeploymentExists reports error if any of the Deployments exists
func ErrorIfDeploymentExists(namespace string, names ...string) error {
	appsCli, err := GetAppsV1Func()
	if err != nil {
		return err
	}
	deployList, err := appsCli.Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil && !kerrs.IsNotFound(err) {
		return err

	}
	for _, d := range deployList.Items {
		for _, n := range names {
			if d.Name == n {
				return fmt.Errorf("existing Deployment %s in namespace %s", d.Name, namespace)
			}
		}
	}
	return nil
}

// ErrorIfServiceExists reports error if any of the Services exists
func ErrorIfServiceExists(namespace string, names ...string) error {
	cli, err := GetCoreV1Func()
	if err != nil {
		return err
	}
	serviceList, err := cli.Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil && !kerrs.IsNotFound(err) {
		return err

	}
	for _, s := range serviceList.Items {
		for _, n := range names {
			if s.Name == n {
				return fmt.Errorf("existing Service %s in namespace %s", s.Name, namespace)
			}
		}
	}
	return nil
}

func setConfigQPSBurst(config *rest.Config) {
	config.Burst = APIServerBurst
	config.QPS = APIServerQPS
}

// GetKubernetesVersion returns the version of Kubernetes cluster in which operator is deployed
func GetKubernetesVersion() (string, error) {
	config, err := GetConfigFromController()
	if err != nil {
		return "", fmt.Errorf("Failed to get kubernetes client config %v", err.Error())
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", fmt.Errorf("Failed to get kubernetes client %v", err.Error())
	}

	versionInfo, err := client.ServerVersion()
	if err != nil {
		return "", fmt.Errorf("Failed to get kubernetes version %v", err.Error())
	}
	return versionInfo.String(), nil
}

func IsMinimumk8sVersion(expectedK8sVersion string) (bool, error) {
	version, err := GetKubernetesVersion()
	if err != nil {
		return false, fmt.Errorf("Failed to get the kubernetes version: %v", err)
	}
	k8sVersion, err := k8sversionutil.ParseSemantic(version)
	if err != nil {
		return false, fmt.Errorf("Failed to parse Kubernetes version %q: %v", k8sVersion, err)
	}
	parsedExpectedK8sVersion := k8sversionutil.MustParseSemantic(expectedK8sVersion)
	if k8sVersion.AtLeast(parsedExpectedK8sVersion) {
		return true, nil
	}
	return false, nil
}