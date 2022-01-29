package categories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/appuio/appuio-cloud-reporting/pkg/erp"
	"github.com/appuio/appuio-cloud-reporting/pkg/erp/entity"
	"github.com/go-logr/logr"
	"github.com/jmoiron/sqlx"
)

// Reconcile synchronizes all stored db.Category with a 3rd party ERP.
// Note: A logger is retrieved from logr.FromContextOrDiscard.
func Reconcile(ctx context.Context, database *sqlx.DB, reconciler erp.CategoryReconciler) error {
	logger := logr.FromContextOrDiscard(ctx).WithName("category")
	logger.Info("Reconciling categories")
	return db.RunInTransaction(ctx, database, func(tx *sqlx.Tx) error {
		var categories []db.Category
		logger.V(2).Info("Retrieving all categories...")
		err := tx.SelectContext(ctx, &categories, "SELECT * FROM categories")
		if err != nil {
			return err
		}
		logger.V(1).Info("Retrieved all categories", "count", len(categories))

		for _, cat := range categories {
			// we need to reconcile categories in the ERP regardless if Target has been set
			logger.V(2).Info("Reconciling category with ERP...", "source", cat.Source)
			modified, err := reconciler.Reconcile(ctx, entity.Category{Source: cat.Source, Target: cat.Target.String})
			if err != nil {
				return fmt.Errorf("error from erp category reconciler: %w", err)
			}
			if cat.Target.Valid {
				// No target update needed
				logger.Info("Category is up-to-date", "source", cat.Source, "target", cat.Target.String)
				continue
			}
			logger.V(2).Info("Updating category...", "id", cat.Id, "source", cat.Source)
			cat.Target = sql.NullString{String: modified.Target, Valid: modified.Target != ""}
			_, err = tx.NamedExecContext(ctx, "UPDATE categories SET target = :target WHERE id = :id", cat)
			if err != nil {
				return err
			}
			logger.Info("Updated category", "source", cat.Source, "target", cat.Target.String)
		}
		logger.Info("Done reconciling categories")
		return nil
	})
}
