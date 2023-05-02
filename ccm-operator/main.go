// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package main

import (
	"flag"
	"github.com/verrazzano/verrazzano-modules/common/controllers/modulelifecycle"
	"github.com/verrazzano/verrazzano-modules/common/controllers/modulelifecycle/handlers/factory"
	moduleplatform "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano/pkg/k8sutil"
	vzlog "github.com/verrazzano/verrazzano/pkg/log"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	controllerruntime "sigs.k8s.io/controller-runtime"
	kzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func main() {
	log := initLog()

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

	// init ccm lifecycle controller
	if err := modulelifecycle.InitController(mgr, factory.NewLifecycleActionHandler(), moduleplatform.CCMLifecycleClass); err != nil {
		log.Errorf("Failed to start CCM controller", err)
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

	utilruntime.Must(moduleplatform.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))

	//+kubebuilder:scaffold:scheme

	return scheme
}

func initLog() *zap.SugaredLogger {
	// Add the zap logger flag set to the CLI.
	opts := kzap.Options{}
	opts.BindFlags(flag.CommandLine)

	flag.Parse()
	kzap.UseFlagOptions(&opts)
	vzlog.InitLogs(opts)

	return zap.S()
}
