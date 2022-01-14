package main

import (
	"fmt"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	dbflag "github.com/appuio/appuio-cloud-reporting/pkg/db/flag"
	"github.com/urfave/cli/v2"
)

type migrateCommand struct {
	ShowPending bool
}

var migrateCommandName = "migrate"

func newMigrateCommand() *cli.Command {
	command := &migrateCommand{}
	return &cli.Command{
		Name:   migrateCommandName,
		Usage:  "Perform database migrations",
		Action: command.execute,
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "show-pending", Usage: "Shows pending migrations and exits", EnvVars: envVars("SHOW_PENDING"), Value: false, Destination: &command.ShowPending},
		},
	}
}

func (cmd *migrateCommand) execute(context *cli.Context) error {
	log := AppLogger(context).WithName(migrateCommandName)
	log.V(1).Info("Opening database connection", "url", dbflag.DatabaseURL)
	rdb, err := db.Open(dbflag.DatabaseURL)
	if err != nil {
		return fmt.Errorf("could not open database connection: %w", err)
	}

	if cmd.ShowPending {
		log.V(1).Info("Showing pending DB migrations")
		pm, err := db.Pending(rdb)
		if err != nil {
			return fmt.Errorf("error showing pending migrations: %w", err)
		}

		for _, p := range pm {
			fmt.Println(p.Name)
		}
		return nil
	}

	log.V(1).Info("Start DB migrations")
	err = db.Migrate(rdb)
	if err != nil {
		return fmt.Errorf("could not migrate database: %w", err)
	}

	return nil
}
