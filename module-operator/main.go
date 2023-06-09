// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package main

import (
	"flag"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/module"
	calicofactory "github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/handlers/calico/factory"
	helmfactory "github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/handlers/helm/factory"
	ccmfactory "github.com/verrazzano/verrazzano-modules/module-operator/controllers/module/handlers/ociccm/factory"
	internalconfig "github.com/verrazzano/verrazzano-modules/module-operator/internal/config"
	"github.com/verrazzano/verrazzano-modules/pkg/k8sutil"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	controllerruntime "sigs.k8s.io/controller-runtime"
	kzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func main() {
	log := initConfig()

	// Use the same controller runtime.Manager for all the controllers, since the manager has the cluster cache and we
	// want the cache to be shared across all the controllers.
	mgr, err := controllerruntime.NewManager(k8sutil.GetConfigOrDieFromController(), controllerruntime.Options{
		Scheme: initScheme(),
		Port:   8080,
	})
	if err != nil {
		log.Errorf("Failed to create a controller-runtime manager", err)
		return
	}

	// init Helm controller
	if err := module.InitController(mgr, helmfactory.NewModuleHandlerInfo(), moduleapi.HelmModuleClass); err != nil {
		log.Errorf("Failed to start Helm controller", err)
		return
	}

	// init Calico controller
	if err := module.InitController(mgr, calicofactory.NewModuleHandlerInfo(), moduleapi.CalicoModuleClass); err != nil {
		log.Errorf("Failed to start the Calico controller", err)
		return
	}

	// init CCM controller
	if err := module.InitController(mgr, ccmfactory.NewModuleHandlerInfo(), moduleapi.CCMModuleClass); err != nil {
		log.Errorf("Failed to start OCI-CCM controller", err)
		return
	}

	// +kubebuilder:scaffold:builder
	log.Info("Starting controller-runtime manager")
	if err := mgr.Start(controllerruntime.SetupSignalHandler()); err != nil {
		log.Errorf("Failed to start controller-runtime manager", err)
		return
	}
}

// initScheme returns the all the schemes used by the controllers.  The controller runtime uses
// a generic go client that can do operations on any object type, but it needs to know the schemes.
func initScheme() *runtime.Scheme {
	// Create a scheme then add each GKV group to the scheme
	scheme := runtime.NewScheme()

	utilruntime.Must(moduleapi.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(appsv1.AddToScheme(scheme))

	//+kubebuilder:scaffold:scheme

	return scheme
}

func initConfig() *zap.SugaredLogger {
	config := internalconfig.Get()

	flag.StringVar(&config.MetricsAddr, "metrics-addr", config.MetricsAddr, "The address the metric endpoint binds to.")
	flag.BoolVar(&config.LeaderElectionEnabled, "enable-leader-election", config.LeaderElectionEnabled,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&config.CertDir, "cert-dir", config.CertDir, "The directory containing tls.crt and tls.key.")

	// Add the zap logger flag set to the CLI.
	opts := kzap.Options{}
	opts.BindFlags(flag.CommandLine)

	flag.Parse()
	kzap.UseFlagOptions(&opts)
	vzlog.InitLogs(opts)

	// Save the config as immutable from this point on.
	internalconfig.Set(config)

	return zap.S()
}
