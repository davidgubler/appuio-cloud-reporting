package main

import (
	"fmt"
	"os"

	"github.com/appuio/appuio-cloud-reporting/cmd/report/report"
)

func main() {
	err := report.Main()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
