package main

import (
	"context"
	"fmt"
	"time"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/appuio/appuio-cloud-reporting/pkg/report"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/api"
	apiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/urfave/cli/v2"
)

type reportCommand struct {
	DatabaseURL   string
	PrometheusURL string
	QueryName     string
	Begin         *time.Time
	RepeatUntil   *time.Time
}

var reportCommandName = "report"

func newReportCommand() *cli.Command {
	command := &reportCommand{}
	return &cli.Command{
		Name:   reportCommandName,
		Usage:  "Run a report for a query in the given period",
		Before: command.before,
		Action: command.execute,
		Flags: []cli.Flag{
			newDbURLFlag(&command.DatabaseURL),
			newPromURLFlag(&command.PrometheusURL),
			&cli.StringFlag{Name: "query-name", Usage: fmt.Sprintf("Name of the query (sample values: %s)", queryNames(db.DefaultQueries)),
				EnvVars: envVars("QUERY_NAME"), Destination: &command.QueryName, Required: true, DefaultText: defaultTestForRequiredFlags},
			&cli.TimestampFlag{Name: "begin", Usage: fmt.Sprintf("Beginning timestamp of the report period in the form of RFC3339 (%s)", time.RFC3339),
				EnvVars: envVars("BEGIN"), Layout: time.RFC3339, Required: true, DefaultText: defaultTestForRequiredFlags},
			&cli.TimestampFlag{Name: "repeat-until", Usage: fmt.Sprintf("Repeat running the report until reaching this timestamp (%s)", time.RFC3339),
				EnvVars: envVars("REPEAT_UNTIL"), Layout: time.RFC3339, Required: false},
		},
	}
}

func (cmd *reportCommand) before(context *cli.Context) error {
	cmd.Begin = context.Timestamp("begin")
	cmd.RepeatUntil = context.Timestamp("repeat-until")
	return LogMetadata(context)
}

func (cmd *reportCommand) execute(cliCtx *cli.Context) error {
	ctx := cliCtx.Context
	log := AppLogger(ctx).WithName(reportCommandName)

	promClient, err := newPrometheusAPIClient(cmd.PrometheusURL)
	if err != nil {
		return fmt.Errorf("could not create prometheus client: %w", err)
	}

	log.V(1).Info("Opening database connection", "url", cmd.DatabaseURL)
	rdb, err := db.Openx(cmd.DatabaseURL)
	if err != nil {
		return fmt.Errorf("could not open database connection: %w", err)
	}
	defer rdb.Close()

	if cmd.RepeatUntil != nil {
		if err := cmd.runReportRange(ctx, rdb, promClient); err != nil {
			return err
		}
	} else {
		if err := cmd.runReport(ctx, rdb, promClient); err != nil {
			return err
		}
	}

	log.Info("Done")
	return nil
}

func (cmd *reportCommand) runReportRange(ctx context.Context, db *sqlx.DB, promClient apiv1.API) error {
	log := AppLogger(ctx)

	log.Info("Running reports...")
	c, err := report.RunRange(db, promClient, cmd.QueryName, *cmd.Begin, *cmd.RepeatUntil)
	log.Info(fmt.Sprintf("Ran %d reports", c))
	return err
}

func (cmd *reportCommand) runReport(ctx context.Context, db *sqlx.DB, promClient apiv1.API) error {
	log := AppLogger(ctx)

	log.V(1).Info("Begin transaction")
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	log.Info("Running report...")
	if err := report.Run(tx, promClient, cmd.QueryName, *cmd.Begin); err != nil {
		return err
	}

	log.V(1).Info("Commit transaction")
	return tx.Commit()
}

func newPrometheusAPIClient(promURL string) (apiv1.API, error) {
	client, err := api.NewClient(api.Config{
		Address: promURL,
	})
	return apiv1.NewAPI(client), err
}
