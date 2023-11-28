package general

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	intentv1 "github.com/5GSEC/nimbus/api/v1"
)

// GeneralController is a struct that holds a Kubernetes client and a WatcherIntent.
type GeneralController struct {
	Client        client.Client  // Client is used to interact with the Kubernetes API.
	WatcherIntent *WatcherIntent // WatcherIntent is a custom struct to manage specific operations.
}

// NewGeneralController creates a new instance of GeneralController.
func NewGeneralController(client client.Client) (*GeneralController, error) {
	if client == nil {
		// If the client is not provided, return an error.
		return nil, fmt.Errorf("GeneralController: Client is nil")
	}

	// Create a new WatcherIntent.
	watcherIntent, err := NewWatcherIntent(client)
	if err != nil {
		// If there is an error in creating WatcherIntent, return an error.
		return nil, fmt.Errorf("GeneralController: Error creating WatcherIntent: %v", err)
	}

	// Return a new GeneralController instance with initialized fields.
	return &GeneralController{
		Client:        client,
		WatcherIntent: watcherIntent,
	}, nil
}

// Reconcile is the method that will be called when there is an update to the resources being watched.
func (gc *GeneralController) Reconcile(ctx context.Context, req ctrl.Request) (*intentv1.SecurityIntent, error) {
	if gc == nil {
		// If the GeneralController instance is nil, return an error.
		return nil, fmt.Errorf("GeneralController is nil")
	}
	if gc.WatcherIntent == nil {
		// If the WatcherIntent is not set, return an error.
		return nil, fmt.Errorf("WatcherIntent is nil")
	}

	// Call the Reconcile method of WatcherIntent to handle the specific logic.
	intent, err := gc.WatcherIntent.Reconcile(ctx, req)
	if err != nil {
		// If there is an error in reconciliation, return the error.
		return nil, fmt.Errorf("Error in WatcherIntent.Reconcile: %v", err)
	}

	// Return the intent and nil as error if reconciliation is successful.
	return intent, nil
}
