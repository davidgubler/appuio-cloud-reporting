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

	categories, err := fetchCategories(ctx, database, logger)
	if err != nil {
		return err
	}

	for _, cat := range categories {
		// We need to reconcile categories in the ERP regardless if Target has been set.
		// These categories in the ERP may have been updated by a 3rd party without the reporting knowing of it.
		// So the reporting being authoritative over categories in the ERP, it should be given chance to reset any changes that deviate from the desired defaults.
		// If we only ever create categories, the categories in the ERP won't ever be touched again.
		logger.V(2).Info("Reconciling category with ERP...", "source", cat.Source)
		input := entity.Category{Source: cat.Source, Target: cat.Target.String}
		output, err := reconciler.Reconcile(ctx, input)
		if err != nil {
			return fmt.Errorf("error from erp category reconciler: %w", err)
		}
		if output == input {
			// No target update needed
			logger.Info("Category is up-to-date", "category", output)
			continue
		}
		err = db.RunInTransaction(ctx, database, func(tx *sqlx.Tx) error {
			logger.V(2).Info("Updating category...", "id", cat.Id, "source", cat.Source)
			cat.Target = sql.NullString{String: output.Target, Valid: output.Target != ""}
			_, err = tx.NamedExecContext(ctx, "UPDATE categories SET target = :target WHERE id = :id", cat)
			if err != nil {
				return err
			}
			logger.Info("Updated category", "source", cat.Source, "target", cat.Target.String)
			return nil
		})
	}
	logger.Info("Done reconciling categories")
	return nil
}

func fetchCategories(ctx context.Context, database *sqlx.DB, logger logr.Logger) ([]db.Category, error) {
	var categories []db.Category
	logger.V(2).Info("Retrieving all categories...")
	err := database.SelectContext(ctx, &categories, "SELECT * FROM categories")
	if err != nil {
		return nil, err
	}
	logger.V(1).Info("Retrieved all categories", "count", len(categories))
	return categories, err
}
