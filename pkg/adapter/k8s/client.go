// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package k8s

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// New returns a new Kubernetes client.
func New(scheme *runtime.Scheme) (client.Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil && errors.Is(err, rest.ErrNotInCluster) {
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig '%v', error: %v", kubeconfig, err)
		}
	}
	k8sClient, err := client.New(config, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create client, error: %v", err)
	}
	return k8sClient, nil
}

// NewDynamicClient returns a Dynamic Kubernetes client.
func NewDynamicClient() dynamic.Interface {
	config, err := rest.InClusterConfig()
	if err != nil && errors.Is(err, rest.ErrNotInCluster) {
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil
		}
	}
	clientSet, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil
	}
	return clientSet
}

// NewOrDie returns a new Kubernetes client and panics if there is an error in
// the config.
func NewOrDie(scheme *runtime.Scheme) client.Client {
	config, err := rest.InClusterConfig()
	if err != nil && errors.Is(err, rest.ErrNotInCluster) {
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(fmt.Sprintf("failed to load kubeconfig '%v', error: %v\n", kubeconfig, err))
		}
	}
	k8sClient, err := client.New(config, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create client, error: %v", err))
	}
	return k8sClient
}

// NewDynamicClientOrDie returns a Dynamic Kubernetes client and panics if there
// is an error in the config.
func NewDynamicClientOrDie() dynamic.Interface {
	config, err := rest.InClusterConfig()
	if err != nil && errors.Is(err, rest.ErrNotInCluster) {
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err)
		}
	}
	clientSet, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return clientSet
}
