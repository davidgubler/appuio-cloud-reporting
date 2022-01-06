package testreport

import (
	"database/sql"
	"flag"
	"fmt"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	dbflag "github.com/appuio/appuio-cloud-reporting/pkg/db/flag"
	"github.com/appuio/appuio-cloud-reporting/pkg/db/types"
)

func Main() error {
	flag.Parse()

	db, err := db.Openx(dbflag.DatabaseURL)
	if err != nil {
		return fmt.Errorf("could not open database connection: %w", err)
	}

	debugCategory := types.Category{
		Source: "debug_category",
		Target: sql.NullString{String: "debug_target", Valid: true},
	}

	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	stmt, err := tx.PrepareNamed("INSERT INTO categories (source, target) VALUES (:source, :target) RETURNING id")
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}

	var id string
	err = stmt.Get(&id, debugCategory)
	if err != nil {
		return fmt.Errorf("error inserting category: %w", err)
	}
	fmt.Println("category has id", id)

	var retreivedCategory types.Category
	err = tx.Get(&retreivedCategory, "SELECT * FROM categories WHERE id=$1", id)
	if err != nil {
		return fmt.Errorf("error retrieving category: %w", err)
	}
	fmt.Println("Category", retreivedCategory)

	tx.Rollback()

	return nil
}
