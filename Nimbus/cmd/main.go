// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	// Importing all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can utilize them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	// Importing custom API types and controllers
	intentv1 "github.com/5GSEC/nimbus/Nimbus/api/v1"
	"github.com/5GSEC/nimbus/Nimbus/controllers"
	cleanup "github.com/5GSEC/nimbus/Nimbus/controllers/cleanup"
	general "github.com/5GSEC/nimbus/Nimbus/controllers/general"
	policy "github.com/5GSEC/nimbus/Nimbus/controllers/policy"

	// Importing third-party Kubernetes resource types
	ciliumv2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	kubearmorhostpolicyv1 "github.com/kubearmor/KubeArmor/pkg/KubeArmorHostPolicy/api/security.kubearmor.com/v1"
	kubearmorpolicyv1 "github.com/kubearmor/KubeArmor/pkg/KubeArmorPolicy/api/security.kubearmor.com/v1"
	//+kubebuilder:scaffold:imports
)

// Global variable for registering schemes.
var (
	scheme   = runtime.NewScheme()        // Scheme registers the API types that the client and server should know.
	setupLog = ctrl.Log.WithName("setup") // Logger specifically for setup.
)

func init() {
	// In init, various Kubernetes and custom resources are added to the scheme.
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(intentv1.AddToScheme(scheme))
	utilruntime.Must(kubearmorpolicyv1.AddToScheme(scheme))
	utilruntime.Must(kubearmorhostpolicyv1.AddToScheme(scheme))
	utilruntime.Must(ciliumv2.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	// Flags for command line parameters such as metrics address, leader election, etc.
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	// Setting the logger with the provided options.
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// Creating a new manager which will manage all the controllers.
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "44502a2e.security.nimbus.com",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "Unable to start manager")
		os.Exit(1)
	}

	// Setting up the GeneralController and PolicyController.
	generalController, err := general.NewGeneralController(mgr.GetClient())
	if err != nil {
		setupLog.Error(err, "Unable to create GeneralController")
		os.Exit(1)
	}

	policyController := policy.NewPolicyController(mgr.GetClient(), mgr.GetScheme())

	// Setting up the SecurityIntentReconciler controller with the manager.
	if err = (&controllers.SecurityIntentReconciler{
		Client:            mgr.GetClient(),
		Scheme:            mgr.GetScheme(),
		GeneralController: generalController,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "SecurityIntent")
		os.Exit(1)
	}

	if err = (&controllers.SecurityIntentBindingReconciler{
		Client:            mgr.GetClient(),
		Scheme:            mgr.GetScheme(),
		GeneralController: generalController,
		PolicyController:  policyController,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "SecurityIntentBinding")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	// Setting up health checks for the manager.
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "Unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "Unable to set up ready check")
		os.Exit(1)
	}

	// Creating channels for handling termination signals.
	sigs := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Separate goroutine to wait for termination signal.
	go func() {
		<-sigs // Waiting for termination signal
		setupLog.Info("Received termination signal, performing cleanup...")

		// Calling the Cleanup function
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		err := cleanup.Cleanup(ctx, mgr.GetClient(), setupLog)

		if err != nil {
			setupLog.Error(err, "Cleanup failed")
		}

		cleanupDone <- true // Signaling cleanup completion
	}()

	// Starting the manager
	setupLog.Info("Starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "Problem running manager")
		os.Exit(1)
	}

	<-cleanupDone // Waiting for cleanup completion
	setupLog.Info("Cleanup completed")
}
