package main

import (
	"fmt"
	"time"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/appuio/appuio-cloud-reporting/pkg/report"
	"github.com/prometheus/client_golang/api"
	apiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/urfave/cli/v2"
)

type reportCommand struct {
	DatabaseURL   string
	PrometheusURL string
	QueryName     string
	Timestamp     *time.Time
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
			&cli.StringFlag{Name: "query-name", Usage: fmt.Sprintf("Name of the query (sample values: %s)", queryNames(db.DefaultQueries)), EnvVars: envVars("QUERY_NAME"), Destination: &command.QueryName, Required: true, DefaultText: defaultTestForRequiredFlags},
			&cli.TimestampFlag{Name: "begin", Usage: fmt.Sprintf("Beginning timestamp of the report period in the form of RFC3339 (%s)", time.RFC3339), EnvVars: envVars("BEGIN"), Layout: time.RFC3339, Required: true, DefaultText: defaultTestForRequiredFlags},
		},
	}
}

func (cmd *reportCommand) before(context *cli.Context) error {
	cmd.Timestamp = context.Timestamp("begin")
	return LogMetadata(context)
}

func (cmd *reportCommand) execute(context *cli.Context) error {
	log := AppLogger(context).WithName(reportCommandName)

	promClient, err := NewPrometheusAPIClient(cmd.PrometheusURL)
	if err != nil {
		return fmt.Errorf("could not create prometheus client: %w", err)
	}

	log.V(1).Info("Opening database connection", "url", cmd.DatabaseURL)
	rdb, err := db.Openx(cmd.DatabaseURL)
	if err != nil {
		return fmt.Errorf("could not open database connection: %w", err)
	}
	defer rdb.Close()

	log.V(1).Info("Begin transaction")
	tx, err := rdb.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	log.Info("Running report...")
	if err := report.Run(tx, promClient, cmd.QueryName, *cmd.Timestamp); err != nil {
		return err
	}

	log.V(1).Info("Commit transaction")
	return tx.Commit()
}

func NewPrometheusAPIClient(promURL string) (apiv1.API, error) {
	client, err := api.NewClient(api.Config{
		Address: promURL,
	})
	return apiv1.NewAPI(client), err
}
