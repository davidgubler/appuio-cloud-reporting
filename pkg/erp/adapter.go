package erp

import (
	"context"
	"fmt"
)

var adapterImpl Adapter

// Adapter abstracts an ERP adapter where the concrete implementation is like a replaceable plugin.
type Adapter interface {
	// Initialize starts and configures the Adapter for consummation.
	Initialize(ctx context.Context) error

	// NewCategoryReconciler returns a new CategoryReconciler instance.
	NewCategoryReconciler() CategoryReconciler

	// Cleanup gracefully shuts down the Adapter.
	Cleanup(ctx context.Context) error
}

// Register registers a new Adapter.
func Register(adapter Adapter) {
	adapterImpl = adapter
}

// Get returns the registered Adapter.
// It returns an error if no Adapter has been registered.
func Get() (Adapter, error) {
	if adapterImpl == nil {
		return nil, fmt.Errorf("no adapter is registered")
	}
	return adapterImpl, nil
}
