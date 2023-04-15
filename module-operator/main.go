// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package main

import (
	"flag"
	platformv1alpha1 "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/controllers/platformctrl/modlifecycle"
	internalconfig "github.com/verrazzano/verrazzano-modules/module-operator/internal/config"
	"os"
	controllerruntime "sigs.k8s.io/controller-runtime"

	modulectrl "github.com/verrazzano/verrazzano-modules/module-operator/controllers/platformctrl/module"
	"github.com/verrazzano/verrazzano/pkg/k8sutil"
	vzlog "github.com/verrazzano/verrazzano/pkg/log"
	"go.uber.org/zap"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	kzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var scheme = runtime.NewScheme()

func init() {
	_ = platformv1alpha1.AddToScheme(scheme)

	// Add K8S api-extensions so that we can list CustomResourceDefinitions during uninstall of VZ
	_ = v1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {

	// config will hold the entire operator config
	config := internalconfig.Get()

	flag.StringVar(&config.MetricsAddr, "metrics-addr", config.MetricsAddr, "The address the metric endpoint binds to.")
	flag.BoolVar(&config.LeaderElectionEnabled, "enable-leader-election", config.LeaderElectionEnabled,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	// Add the zap logger flag set to the CLI.
	opts := kzap.Options{}
	opts.BindFlags(flag.CommandLine)

	flag.Parse()
	kzap.UseFlagOptions(&opts)
	//vzlog.InitLogs(opts)

	// Save the config as immutable from this point on.
	log := zap.S()

	mgr, err := controllerruntime.NewManager(k8sutil.GetConfigOrDieFromController(), controllerruntime.Options{
		Scheme:             scheme,
		MetricsBindAddress: config.MetricsAddr,
		Port:               8080,
		LeaderElection:     config.LeaderElectionEnabled,
		LeaderElectionID:   "3ec4d290.verrazzano.io",
	})
	if err != nil {
		log.Errorf("Failed to create a controller-runtime manager: %v", err)
		os.Exit(1)
	}

	log.Info("Starting Verrazzano Module Operator")

	if err := (&modlifecycle.Reconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "Failed to setup ModuleLifecycle controller", vzlog.FieldController, "ModuleLifecycleController")
		os.Exit(1)
	}

	// v1beta2 VerrazzanoModule controller
	if err = (&modulectrl.VerrazzanoModuleReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "Failed to setup Verrazzano Module controller", vzlog.FieldController, "VerrazzanoModuleController")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder
	log.Info("Starting controller-runtime manager")
	if err := mgr.Start(controllerruntime.SetupSignalHandler()); err != nil {
		log.Errorf("Failed starting controller-runtime manager: %v", err)
		os.Exit(1)
	}
}
