// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package k8s

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	kspAPI "github.com/kubearmor/KubeArmor/pkg/KubeArmorController/api/security.kubearmor.com/v1"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// NewClient creates a new Kubernetes client to perform CRUD operations.
func NewClient(logger *zap.SugaredLogger) client.Client {
	err := registerScheme()
	if err != nil {
		logger.Errorf("failed to register scheme, error: %v", err)
	}
	config, err := rest.InClusterConfig()
	if err != nil && errors.Is(err, rest.ErrNotInCluster) {
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			logger.Errorf("failed to load kubeconfig, error: %v", err)
		}
	}
	k8sClient, err := client.New(config, client.Options{})
	if err != nil {
		logger.Fatalf("failed to create client, error: %v", err)
	}

	// Temporary fix for 👇 error
	// [controller-runtime] log.SetLogger(...) was never called; logs will not be displayed.
	log.SetLogger(logr.Logger{})
	return k8sClient
}

func registerScheme() error {
	// Add a new scheme when adding a new Adapter
	err := kspAPI.AddToScheme(scheme.Scheme)
	return err
}
