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

type ClusterNimbusPolicy struct {
	Client client.Client
}

func NewClusterNimbusPolicy(client client.Client) (*ClusterNimbusPolicy, error) {
	if client == nil {
		return nil, fmt.Errorf("ClusterNimbusPolicyWatcher: Client is nil")
	}
	return &ClusterNimbusPolicy{
		Client: client,
	}, nil
}

func (wcnp *ClusterNimbusPolicy) Reconcile(ctx context.Context, req ctrl.Request) (*v1.ClusterNimbusPolicy, error) {
	logger := log.FromContext(ctx)
	if wcnp == nil || wcnp.Client == nil {
		logger.Info("ClusterNimbusPolicy watcher is nil or Client is nil")
		return nil, fmt.Errorf("ClusterNimbusPolicy watcher or Client is not initialized")
	}

	var cwnp v1.ClusterNimbusPolicy
	err := wcnp.Client.Get(ctx, types.NamespacedName{Name: req.Name}, &cwnp)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("ClusterNimbusPolicy resource not found. Ignoring since object must be deleted", "Name", req.Name)
			return nil, nil
		}
		logger.Error(err, "failed to get ClusterNimbusPolicy", "Name", req.Name)
		return nil, err
	}
	return &cwnp, nil
}
