package check

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// MissingField represents a missing field.
type MissingField struct {
	Table string

	ID     string
	Source string

	MissingField string
}

const missingQuery = `
	SELECT 'categories' as table, id, source, 'target' as missingfield FROM categories WHERE target IS NULL OR target = ''
	UNION ALL
	SELECT 'tenants' as table, id, source, 'target' as missingfield FROM tenants WHERE target IS NULL OR target = ''
	UNION ALL
	SELECT 'products' as table, id, source, 'target' as missingfield FROM products WHERE target IS NULL OR target = ''
	UNION ALL
	SELECT 'products' as table, id, source, 'amount' as missingfield FROM products WHERE amount = 0
	UNION ALL
	SELECT 'products' as table, id, source, 'unit' as missingfield FROM products WHERE unit = ''
`

// Missing checks for missing fields in the reporting database.
func Missing(ctx context.Context, tx sqlx.QueryerContext) ([]MissingField, error) {
	var missing []MissingField

	err := sqlx.SelectContext(ctx, tx, &missing, fmt.Sprintf(`WITH missing AS (%s) SELECT * FROM missing ORDER BY "table",missingfield,source`, missingQuery))
	return missing, err
}
