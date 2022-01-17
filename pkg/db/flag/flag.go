package dbflag

import (
	"flag"
	"os"
)

var DatabaseURL string

func init() {
	flag.StringVar(&DatabaseURL, "db-url", urlFromEnv(), "The URL to connect to the database.")
}

func urlFromEnv() string {
	if url, exists := os.LookupEnv("DB_URL"); exists {
		return url
	}
	return "postgres://postgres@localhost/reporting-db?sslmode=disable"
}
