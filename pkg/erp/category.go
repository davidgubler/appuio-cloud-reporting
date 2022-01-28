package erp

import (
	"context"

	"github.com/appuio/appuio-cloud-reporting/pkg/erp/entity"
)

// CategoryReconciler reconciles entity.Category instances.
type CategoryReconciler interface {
	// Reconcile takes the given category and reconciles it with the concrete ERP implementation.
	// The CategoryReconciler may return a modified entity.Category instance or the same one if there were no changes.
	// An error is returned if reconciliation failed.
	Reconcile(ctx context.Context, category entity.Category) (entity.Category, error)
}
