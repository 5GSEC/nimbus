// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package controller

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/5GSEC/nimbus/api/v1"
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

func extractBoundIntentsInfo(intents []v1.MatchIntent) (int32, []string) {
	var count int32
	var names []string
	for _, intent := range intents {
		count++
		names = append(names, intent.Name)
	}
	return count, names
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
	case *v1.NimbusPolicy:
		objToGet = &v1.SecurityIntentBinding{}
	case *v1.ClusterNimbusPolicy:
		objToGet = &v1.ClusterSecurityIntentBinding{}
	}

	if err := c.Get(context.Background(), types.NamespacedName{Name: ownerName, Namespace: controllee.GetNamespace()}, objToGet); err != nil {
		return false
	}

	// Verify whether the controller we found is same that the ControllerRef points
	// to.
	return objToGet.GetUID() == ownerUid
}
