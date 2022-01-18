package main

import (
	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/urfave/cli/v2"
)

const defaultTestForRequiredFlags = "<required>"

func newDbURLFlag(destination *string) *cli.StringFlag {
	return &cli.StringFlag{Name: "db-url", Usage: "Database connection URL in the form of postgres://user@host:port/db-name?option=value",
		EnvVars: envVars("DB_URL"), Destination: destination, Required: true, DefaultText: defaultTestForRequiredFlags}
}

func newPromURLFlag(destination *string) *cli.StringFlag {
	return &cli.StringFlag{Name: "prom-url", Usage: "Prometheus connection URL in the form of http://host:port",
		EnvVars: envVars("PROM_URL"), Destination: destination, Value: "http://localhost:9090"}
}

func queryNames(queries []db.Query) []string {
	names := make([]string, len(queries))
	for i := range queries {
		names[i] = queries[i].Name
	}
	return names
}
