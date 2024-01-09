// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package kubearmor

import (
	"context"
	"fmt"
	"strings"

	kubearmorv1 "github.com/kubearmor/KubeArmor/pkg/KubeArmorController/api/security.kubearmor.com/v1"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/5GSEC/nimbus/api/v1"
	"github.com/5GSEC/nimbus/pkg/adapter/idpool"
	"github.com/5GSEC/nimbus/pkg/adapter/kubearmor/processor"
)

type Client struct {
	Logger    *zap.SugaredLogger
	k8sClient client.Client
}

func NewKubeArmorClient(logger *zap.SugaredLogger, client client.Client) *Client {
	return &Client{
		Logger:    logger,
		k8sClient: client,
	}
}

func (c *Client) ApplyPolicy(ctx context.Context, np v1.NimbusPolicy) error {
	// Build KSPs based on given IDs
	var ksps []kubearmorv1.KubeArmorPolicy
	for _, nimbusRule := range np.Spec.NimbusRules {
		id := nimbusRule.Id
		if idpool.IdSupportedBy(id, "kubearmor") {
			ksp := processor.BuildKspFor(id)
			ksp.Name = np.Name + "-" + strings.ToLower(id)
			ksp.Namespace = np.Namespace
			ksp.Spec.Message = nimbusRule.Description
			ksp.Spec.Selector.MatchLabels = np.Spec.Selector.MatchLabels
			ksp.Spec.Action = kubearmorv1.ActionType(nimbusRule.Rule.RuleAction)
			processor.ProcessRuleParams(&ksp, nimbusRule.Rule)
			ksps = append(ksps, ksp)
		} else {
			c.Logger.Warnf("KubeArmor does not support '%s' ID", id)
		}
	}

	// Iterate using a separate index variable to avoid aliasing
	for idx := range ksps {
		ksp := ksps[idx]
		var existingKsp kubearmorv1.KubeArmorPolicy
		err := c.k8sClient.Get(ctx, types.NamespacedName{Name: ksp.Name, Namespace: ksp.Namespace}, &existingKsp)
		if err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("failed to get existing '%s' KubeArmorPolicy, error: %v", ksp.Name, err)
		}
		if errors.IsNotFound(err) {
			if err = c.k8sClient.Create(ctx, &ksp); err != nil {
				return fmt.Errorf("failed to create '%s' KubeArmorPolicy, error: %v", ksp.Name, err)
			}
			c.Logger.Infof("Created KubeArmorPolicy %s", ksp.Name)
		} else {
			if err = c.k8sClient.Update(ctx, &ksp); err != nil {
				return fmt.Errorf("failed to apply existing '%s' KubeArmorPolicy, error: %v", ksp.Name, err)
			}
			c.Logger.Infof("Configured KubeArmorPolicy %s", ksp.Name)
		}
	}
	return nil
}

func (c *Client) DeletePolicy(ctx context.Context, np v1.NimbusPolicy) error {
	for _, nimbusRule := range np.Spec.NimbusRules {
		id := nimbusRule.Id
		if idpool.IdSupportedBy(id, "kubearmor") {
			var ksp kubearmorv1.KubeArmorPolicy
			ksp.SetName(np.Name + "-" + strings.ToLower(id))
			ksp.SetNamespace(np.Namespace)
			if err := c.k8sClient.Delete(ctx, &ksp); err != nil && !errors.IsNotFound(err) {
				return fmt.Errorf("failed to delete '%s' KubeArmorPolicy, error: %v", ksp.Name, err)
			}
			c.Logger.Infof("Deleted KubeArmorPolicy %s", ksp.Name)
		}
	}
	return nil
}
