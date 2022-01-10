package main

import (
	"fmt"
	"os"

	"github.com/appuio/appuio-cloud-reporting/cmd/testreport/testreport"
)

func main() {
	err := testreport.Main()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
