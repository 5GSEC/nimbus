package general

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	intentv1 "github.com/5GSEC/nimbus/api/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

// WatcherIntent is a struct that holds a Kubernetes client.
type WatcherIntent struct {
	Client client.Client // Client to interact with Kubernetes resources.
}

// NewWatcherIntent creates a new instance of WatcherIntent.
func NewWatcherIntent(client client.Client) (*WatcherIntent, error) {
	if client == nil {
		// Return an error if the client is not provided.
		return nil, fmt.Errorf("WatcherIntent: Client is nil")
	}

	// Return a new WatcherIntent instance with the provided client.
	return &WatcherIntent{
		Client: client,
	}, nil
}

// Reconcile is the method that handles the reconciliation of the Kubernetes resources.
func (wi *WatcherIntent) Reconcile(ctx context.Context, req ctrl.Request) (*intentv1.SecurityIntent, error) {
	log := log.FromContext(ctx) // Get the logger from the context.

	// Check if WatcherIntent or its client is not initialized.
	if wi == nil || wi.Client == nil {
		fmt.Println("WatcherIntent is nil or Client is nil in Reconcile")
		return nil, fmt.Errorf("WatcherIntent or Client is not initialized")
	}

	intent := &intentv1.SecurityIntent{} // Create an instance of SecurityIntent.
	// Attempt to get the SecurityIntent resource from Kubernetes.
	err := wi.Client.Get(ctx, types.NamespacedName{
		Name:      req.Name,
		Namespace: req.Namespace,
	}, intent)

	if err != nil {
		// Handle the case where the SecurityIntent resource is not found.
		if errors.IsNotFound(err) {
			log.Info("SecurityIntent resource not found. Ignoring since object must be deleted")
			return nil, nil
		}
		// Log and return an error if there is a problem getting the SecurityIntent.
		log.Error(err, "Failed to get SecurityIntent")
		return nil, err
	}

	// Return the SecurityIntent instance if found successfully.
	return intent, nil
}
