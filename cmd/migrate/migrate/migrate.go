package migrate

import (
	"flag"
	"fmt"

	"github.com/appuio/appuio-public-reporting/pkg/db"
	dbflag "github.com/appuio/appuio-public-reporting/pkg/db/flag"
	"github.com/appuio/appuio-public-reporting/pkg/db/migrations"
)

func Main() error {
	var showPending bool
	flag.BoolVar(&showPending, "show-pending", false, "Shows pending migrations if set.")
	flag.Parse()

	db, err := db.Open(dbflag.DatabaseURL)
	if err != nil {
		return fmt.Errorf("could not open database connection: %w", err)
	}

	if showPending {
		pm, err := migrations.Pending(db)
		if err != nil {
			return fmt.Errorf("error showing pending migrations: %w", err)
		}

		for _, p := range pm {
			fmt.Println(p.Name)
		}
		return nil
	}

	err = migrations.Migrate(db)
	if err != nil {
		return fmt.Errorf("could not migrate database: %w", err)
	}

	return nil
}
