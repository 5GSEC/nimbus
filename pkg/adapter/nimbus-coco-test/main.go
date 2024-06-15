// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/5GSEC/nimbus/pkg/adapter/nimbus-coco/manager"
	pb "github.com/5GSEC/nimbus/pkg/grpc"
	"github.com/go-logr/logr"
)

type server struct {
	pb.UnimplementedResourceDataServiceServer
	resourceDataCh chan *pb.ResourceData
}

func (s *server) SendPodData(ctx context.Context, in *pb.ResourceData) (*pb.Response, error) {
	logger := log.FromContext(ctx)
	logger.Info("Received resource data", "Resource.Name", in.Name, "Resource.Namespace", in.Namespace)
	s.resourceDataCh <- in
	return &pb.Response{Message: "Resource data received"}, nil
}

func startGRPCServer(logger logr.Logger, resourceDataCh chan *pb.ResourceData) {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	pb.RegisterResourceDataServiceServer(s, &server{resourceDataCh: resourceDataCh})

	logger.Info("gRPC server on port 50051 started")

	if err := s.Serve(lis); err != nil {
		panic(err)
	}
}

func main() {
	ctrl.SetLogger(zap.New())
	logger := ctrl.Log

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

	resourceDataCh := make(chan *pb.ResourceData)
	go startGRPCServer(logger, resourceDataCh)

	manager.Run(ctx, resourceDataCh)
	logger.Info("Coco adapter started")
}
