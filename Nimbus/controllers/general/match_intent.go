// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package general

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/log"

	intentv1 "github.com/5GSEC/nimbus/Nimbus/api/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// BindingInfo holds the matching information between SecurityIntent and SecurityIntentBinding.
type BindingInfo struct {
	Intent  []*intentv1.SecurityIntent
	Binding *intentv1.SecurityIntentBinding
}

// MatchIntentAndBinding finds matching SecurityIntent for a given SecurityIntentBinding.
func MatchIntentAndBinding(ctx context.Context, client client.Client, binding *intentv1.SecurityIntentBinding) (*BindingInfo, error) {
	log := log.FromContext(ctx)

	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}
	if binding == nil {
		return nil, fmt.Errorf("SecurityIntentBinding is nil")
	}

	var matchedIntents []*intentv1.SecurityIntent
	for _, intentReq := range binding.Spec.IntentRequests {
		intent := &intentv1.SecurityIntent{}
		err := client.Get(ctx, types.NamespacedName{Name: intentReq.IntentName, Namespace: binding.Namespace}, intent)
		if err != nil {
			log.Error(err, "Failed to get SecurityIntent", "IntentName", intentReq.IntentName, "Namespace", binding.Namespace)
			return nil, fmt.Errorf("Failed to get SecurityIntent '%s' in namespace '%s': %v", intentReq.IntentName, binding.Namespace, err)
		}
		matchedIntents = append(matchedIntents, intent)
	}

	if len(matchedIntents) > 0 {
		log.Info("Matched SecurityIntents for Binding", "BindingName", binding.Name)
	}
	return &BindingInfo{
		Intent:  matchedIntents,
		Binding: binding,
	}, nil
}
