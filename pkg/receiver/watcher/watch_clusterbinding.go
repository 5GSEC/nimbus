package watcher

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/5GSEC/nimbus/api/v1"
)

type WatcherClusterBinding struct {
	Client client.Client
}

func NewWatcherClusterBinding(client client.Client) (*WatcherClusterBinding, error) {
	if client == nil {
		return nil, fmt.Errorf("WatcherClusterBinding: Client is nil")
	}

	// Return a new WatcherBinding instance with the provided client.
	return &WatcherClusterBinding{
		Client: client,
	}, nil
}

func (wcb *WatcherClusterBinding) Reconcile(ctx context.Context, req ctrl.Request) (*v1.ClusterSecurityIntentBinding, error) {
	logger := log.FromContext(ctx)
	if wcb == nil || wcb.Client == nil {
		logger.Info("WatcherClusterBinding is nil or Client is nil in Reconcile")
		return nil, fmt.Errorf("WatcherClusterBinding or Client is not initialized")
	}

	clusterSib := &v1.ClusterSecurityIntentBinding{}
	err := wcb.Client.Get(ctx, types.NamespacedName{Name: req.Name}, clusterSib)

	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		logger.Error(err, "failed to get ClusterSecurityIntentBinding", "Name", req.Name)
		return nil, err
	}
	return clusterSib, nil
}
