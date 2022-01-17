package main

import (
	"database/sql"
	"fmt"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/urfave/cli/v2"
)

type testReportCommand struct {
	DatabaseURL string
}

var testReportCommandName = "testreport"

func newTestReportCommand() *cli.Command {
	command := &testReportCommand{}
	return &cli.Command{
		Name:   testReportCommandName,
		Usage:  "For quickly testing something in local development",
		Before: command.before,
		Action: command.execute,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "db-url", Usage: "Database connection URL in the form of postgres://user@host:port/db-name?option=value", EnvVars: envVars("DB_URL"), Destination: &command.DatabaseURL, Required: true},
		},
	}
}

func (cmd *testReportCommand) before(ctx *cli.Context) error {
	return logMetadata(ctx)
}

func (cmd *testReportCommand) execute(context *cli.Context) error {
	log := AppLogger(context).WithName(migrateCommandName)
	log.V(1).Info("Opening database connection", "url", cmd.DatabaseURL)

	rdb, err := db.Openx(cmd.DatabaseURL)
	if err != nil {
		return fmt.Errorf("could not open database connection: %w", err)
	}
	defer rdb.Close()

	debugCategory := db.Category{
		Source: "debug_category",
		Target: sql.NullString{String: "debug_target", Valid: true},
	}

	log.V(1).Info("Begin transaction")
	tx, err := rdb.Beginx()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	log.V(1).Info("Prepare Query")
	stmt, err := tx.PrepareNamed("INSERT INTO categories (source, target) VALUES (:source, :target) RETURNING id")
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	log.V(1).Info("Insert data")
	var id string
	err = stmt.Get(&id, debugCategory)
	if err != nil {
		return fmt.Errorf("error inserting category: %w", err)
	}
	log.Info("Retrieved Category ID", "id", id)

	log.V(1).Info("Select category")
	var retreivedCategory db.Category
	err = tx.Get(&retreivedCategory, "SELECT * FROM categories WHERE id=$1", id)
	if err != nil {
		return fmt.Errorf("error retrieving category: %w", err)
	}
	log.Info("Retrieved Category", "category", retreivedCategory)

	return nil
}
