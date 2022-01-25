package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/appuio/appuio-cloud-reporting/pkg/invoice"
)

type invoiceCommand struct {
	DatabaseURL string
	Year        int
	Month       time.Month
}

var invoiceCommandName = "invoice"

func newInvoiceCommand() *cli.Command {
	command := &invoiceCommand{}
	return &cli.Command{
		Name:   invoiceCommandName,
		Usage:  "Run a invoice for a query in the given period",
		Before: command.before,
		Action: command.execute,
		Flags: []cli.Flag{
			newDbURLFlag(&command.DatabaseURL),

			&cli.IntFlag{Name: "year", Usage: "Year to generate the report for.",
				EnvVars: envVars("YEAR"), Destination: &command.Year, Required: true},
			&cli.IntFlag{Name: "month", Usage: "Month to generate the report for.",
				EnvVars: envVars("MONTH"), Destination: (*int)(&command.Month), Required: true},
		},
	}
}

func (cmd *invoiceCommand) before(context *cli.Context) error {
	if cmd.Month < 1 || cmd.Month > 12 {
		return fmt.Errorf("unknown month %q", cmd.Month)
	}
	return LogMetadata(context)
}

func (cmd *invoiceCommand) execute(cliCtx *cli.Context) error {
	ctx := cliCtx.Context
	log := AppLogger(ctx).WithName(invoiceCommandName)

	log.V(1).Info("Opening database connection", "url", cmd.DatabaseURL)
	rdb, err := db.Openx(cmd.DatabaseURL)
	if err != nil {
		return fmt.Errorf("could not open database connection: %w", err)
	}
	defer rdb.Close()

	log.V(1).Info("Begin transaction")
	tx, err := rdb.BeginTxx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	invoices, err := invoice.Generate(ctx, tx, cmd.Year, cmd.Month)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "\t")
	enc.Encode(invoices)
	if err := enc.Encode(invoices); err != nil {
		return err
	}

	return tx.Commit()
}
