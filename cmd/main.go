// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package main

import (
	"flag"
	"github.com/5GSEC/nimbus/pkg/util"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	v1 "github.com/5GSEC/nimbus/api/v1alpha1"
	"github.com/5GSEC/nimbus/internal/controller"
	// Importing third-party Kubernetes resource types
	//+kubebuilder:scaffold:imports
)

// Global variables for scheme registration and setup logging.
var (
	scheme   = k8sruntime.NewScheme()     // Scheme for registering API types for client and server.
	setupLog = ctrl.Log.WithName("setup") // Logger for setup process.
)

func init() {
	// In init, various Kubernetes and custom resources are added to the scheme.
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	// Flags for command line parameters such as metrics address, leader election, etc.
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	recoverPanic := true

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	// Setting the logger with the provided options.
	ctrl.SetLogger(zap.New())
	util.LogBuildInfo(ctrl.Log)

	// Creating a new manager which will manage all the controllers.
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress: probeAddr,
		Controller: config.Controller{
			RecoverPanic: &recoverPanic,
		},
		LeaderElection:   enableLeaderElection,
		LeaderElectionID: "44502a2e.security.nimbus.com",
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

	if err = (&controller.SecurityIntentReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "SecurityIntent")
		os.Exit(1)
	}

	if err = (&controller.SecurityIntentBindingReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "SecurityIntentBinding")
		os.Exit(1)
	}

	if err = (&controller.ClusterSecurityIntentBindingReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ClusterSecurityIntentBinding")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	// Adding health and readiness checks.
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "Unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "Unable to set up ready check")
		os.Exit(1)
	}

	// Starting the controller manager.
	setupLog.Info("Starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "Problem running manager")
		os.Exit(1)
	}
}
