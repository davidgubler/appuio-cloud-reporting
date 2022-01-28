package categories

import (
	"context"
	"database/sql"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/appuio/appuio-cloud-reporting/pkg/erp"
	"github.com/appuio/appuio-cloud-reporting/pkg/erp/entity"
	"github.com/jmoiron/sqlx"
)

// Reconcile synchronizes all stored db.Category with a 3rd party ERP.
func Reconcile(ctx context.Context, database *sqlx.DB, reconciler erp.CategoryReconciler) error {
	return db.RunInTransaction(ctx, database, func(tx *sqlx.Tx) error {
		var categories []db.Category
		err := tx.SelectContext(ctx, &categories, "SELECT * FROM categories")
		if err != nil {
			return err
		}

		for _, cat := range categories {
			// we need to reconcile categories in the ERP regardless if Target has been set
			modified, err := reconciler.Reconcile(ctx, entity.Category{Source: cat.Source, Target: cat.Target.String})
			if err != nil {
				return err
			}
			if cat.Target.Valid {
				// No target update needed
				continue
			}
			cat.Target = sql.NullString{String: modified.Target, Valid: modified.Target != ""}
			_, err = tx.NamedExecContext(ctx, "UPDATE categories SET target = :target WHERE id = :id", cat)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
