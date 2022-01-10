package migrate

import (
	"flag"
	"fmt"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	dbflag "github.com/appuio/appuio-cloud-reporting/pkg/db/flag"
)

func Main() error {
	var showPending, seed bool
	flag.BoolVar(&seed, "seed", false, "Seeds the database.")
	flag.BoolVar(&showPending, "show-pending", false, "Shows pending migrations if set.")
	flag.Parse()

	rdb, err := db.Open(dbflag.DatabaseURL)
	if err != nil {
		return fmt.Errorf("could not open database connection: %w", err)
	}

	if seed {
		err := db.Seed(rdb)
		if err != nil {
			return fmt.Errorf("error seeding database: %w", err)
		}

		fmt.Println("Done seeding...")
		return nil
	}

	if showPending {
		pm, err := db.Pending(rdb)
		if err != nil {
			return fmt.Errorf("error showing pending migrations: %w", err)
		}

		for _, p := range pm {
			fmt.Println(p.Name)
		}
		return nil
	}

	err = db.Migrate(rdb)
	if err != nil {
		return fmt.Errorf("could not migrate database: %w", err)
	}

	return nil
}
