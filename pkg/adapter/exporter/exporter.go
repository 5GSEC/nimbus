// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package exporter

import (
	"context"

	"go.uber.org/zap"

	v1 "github.com/5GSEC/nimbus/api/v1"
	"github.com/5GSEC/nimbus/pkg/adapter"
	"github.com/5GSEC/nimbus/pkg/adapter/k8s"
	"github.com/5GSEC/nimbus/pkg/adapter/kubearmor"
)

// ExportNpToAdapters export nimbus policy to security-engine adapters.
func ExportNpToAdapters(loggr *zap.SugaredLogger, nimbusPolicy v1.NimbusPolicy) {
	for _, adptr := range adapter.Adapters {
		loggr.Infof("Exporting '%s' NimbusPolicy to %s security engine", nimbusPolicy.Name, adptr)
		err := sendNpTo(loggr, nimbusPolicy, adptr)
		if err != nil {
			loggr.Warnf("%v", err)
		}
	}
}

func sendNpTo(loggr *zap.SugaredLogger, nimbusPolicy v1.NimbusPolicy, adptr string) error {
	var securityEngineClient adapter.Adapter
	k8sClient := k8s.NewClient(loggr)
	switch adptr {
	case "kubearmor":
		securityEngineClient = kubearmor.NewKubeArmorClient(loggr, k8sClient)
		err := securityEngineClient.ApplyPolicy(context.Background(), nimbusPolicy)
		return err
	default:
		return nil
	}
}
