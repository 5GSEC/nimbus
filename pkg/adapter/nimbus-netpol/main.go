// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/5GSEC/nimbus/pkg/adapter/nimbus-netpol/manager"
	"github.com/5GSEC/nimbus/pkg/adapter/watcher"
)

func main() {
	ctrl.SetLogger(zap.New())
	logger := ctrl.Log

	ctx, cancelFunc := context.WithCancel(context.Background())
	ctrl.LoggerInto(ctx, logger)

	nimbusPolicyCh := make(chan [2]string)
	nimbusPolicyToDeleteCh := make(chan [2]string)
	nimbusPolicyUpdateCh := make(chan [2]string)
	go watcher.WatchNimbusPolicies(ctx, nimbusPolicyCh, nimbusPolicyToDeleteCh, nimbusPolicyUpdateCh)

	clusterNpChan := make(chan string)
	clusterNpToDeleteChan := make(chan string)
	go watcher.WatchClusterNimbusPolicies(ctx, clusterNpChan, clusterNpToDeleteChan)

	go func() {
		termChan := make(chan os.Signal)
		signal.Notify(termChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-termChan
		logger.Info("Shutdown signal received, waiting for all workers to finish")
		cancelFunc()
		logger.Info("All workers finished, shutting down")
	}()

	logger.Info("Network Policy adapter started")
	manager.ManageNetPols(ctx, nimbusPolicyCh, nimbusPolicyToDeleteCh, clusterNpChan, clusterNpToDeleteChan)
}
