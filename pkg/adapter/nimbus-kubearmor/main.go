// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package main

import (
	"context"
	"github.com/5GSEC/nimbus/pkg/util"
	"os"
	"os/signal"
	"syscall"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/5GSEC/nimbus/pkg/adapter/nimbus-kubearmor/manager"
)

func main() {
	ctrl.SetLogger(zap.New())
	logger := ctrl.Log
	util.LogBuildInfo(logger)

	ctx, cancelFunc := context.WithCancel(context.Background())
	ctrl.LoggerInto(ctx, logger)

	go func() {
		termChan := make(chan os.Signal)
		signal.Notify(termChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-termChan
		logger.Info("Shutdown signal received, waiting for all workers to finish")
		cancelFunc()
		logger.Info("All workers finished, shutting down")
	}()

	logger.Info("KubeArmor adapter started")
	manager.Run(ctx)
}
