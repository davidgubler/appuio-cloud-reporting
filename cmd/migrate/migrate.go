package main

import (
	"fmt"
	"os"

	"github.com/appuio/appuio-public-invoicing/cmd/migrate/migrate"
)

func main() {
	err := migrate.Main()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
