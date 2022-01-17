package report

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	dbflag "github.com/appuio/appuio-cloud-reporting/pkg/db/flag"
	"github.com/appuio/appuio-cloud-reporting/pkg/promflag"
	"github.com/appuio/appuio-cloud-reporting/pkg/report"
)

func Main() error {
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s QUERY_NAME TIMESTAMP_RFC3339\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(3)
	}

	queryName := args[0]
	ts, err := time.Parse(time.RFC3339, args[1])
	if err != nil {
		return fmt.Errorf("could not parse given timestamp: %w", err)
	}

	promClient, err := promflag.PrometheusAPIClient()
	if err != nil {
		return fmt.Errorf("could not create prometheus client: %w", err)
	}

	rdb, err := db.Openx(dbflag.DatabaseURL)
	if err != nil {
		return fmt.Errorf("could not open database connection: %w", err)
	}
	defer rdb.Close()

	tx, err := rdb.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := report.Run(tx, promClient, queryName, ts); err != nil {
		return err
	}

	return tx.Commit()
}
