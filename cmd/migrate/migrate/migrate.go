package migrate

import (
	"flag"
	"fmt"

	"github.com/appuio/appuio-public-invoicing/pkg/db"
	dbflag "github.com/appuio/appuio-public-invoicing/pkg/db/flag"
	"github.com/appuio/appuio-public-invoicing/pkg/db/migrations"
)

func Main() error {
	flag.Parse()

	db, err := db.Open(dbflag.DatabaseURL)
	if err != nil {
		return fmt.Errorf("could not open database connection: %w", err)
	}

	err = migrations.Migrate(db)
	if err != nil {
		return fmt.Errorf("could not migrate database: %w", err)
	}

	return nil
}
