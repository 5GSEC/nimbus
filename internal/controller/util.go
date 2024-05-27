// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package controller

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1alpha1 "github.com/5GSEC/nimbus/api/v1alpha1"
)

// TODO: Add constants for recommend labels and update objects accordingly.
// https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/

const (
	StatusCreated = "Created"
)

func doNotRequeue() (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func requeueWithError(err error) (ctrl.Result, error) {
	return ctrl.Result{}, err
}

func extractBoundIntentsNameFromSib(ctx context.Context, c client.Client, name, namespace string) []string {
	logger := log.FromContext(ctx)

	var boundIntentsName []string

	var sib v1alpha1.SecurityIntentBinding
	if err := c.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &sib); err != nil {
		logger.Error(err, "failed to fetch SecurityIntentBinding", "securityIntentBindingName", name, "securityIntentBindingNamespace", namespace)
		return boundIntentsName
	}

	for _, intent := range sib.Spec.Intents {
		var si v1alpha1.SecurityIntent
		if err := c.Get(ctx, types.NamespacedName{Name: intent.Name}, &si); err == nil {
			boundIntentsName = append(boundIntentsName, intent.Name)
		}
	}

	return boundIntentsName
}
func extractBoundIntentsNameFromCSib(ctx context.Context, c client.Client, name string) []string {
	logger := log.FromContext(ctx)

	var boundIntentsName []string

	var csib v1alpha1.ClusterSecurityIntentBinding
	if err := c.Get(ctx, types.NamespacedName{Name: name}, &csib); err != nil {
		logger.Error(err, "failed to fetch ClusterSecurityIntentBinding", "ClusterSecurityIntentBinding", name)
		return boundIntentsName
	}

	for _, intent := range csib.Spec.Intents {
		var si v1alpha1.SecurityIntent
		if err := c.Get(ctx, types.NamespacedName{Name: intent.Name}, &si); err == nil {
			boundIntentsName = append(boundIntentsName, intent.Name)
		}
	}

	return boundIntentsName
}

func ownerExists(c client.Client, controllee client.Object) bool {
	// Don't even try to look if it has no ControllerRef.
	controller := metav1.GetControllerOf(controllee)
	if controller == nil {
		return false
	}

	ownerName := controller.Name
	ownerUid := controller.UID
	var objToGet client.Object

	switch controllee.(type) {
	case *v1alpha1.NimbusPolicy:
		objToGet = &v1alpha1.SecurityIntentBinding{}
	case *v1alpha1.ClusterNimbusPolicy:
		objToGet = &v1alpha1.ClusterSecurityIntentBinding{}
	}

	if err := c.Get(context.Background(), types.NamespacedName{Name: ownerName, Namespace: controllee.GetNamespace()}, objToGet); err != nil {
		return false
	}

	// Verify whether the controller we found is same that the ControllerRef points
	// to.
	return objToGet.GetUID() == ownerUid
}
