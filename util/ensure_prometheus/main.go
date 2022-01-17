package main

import (
	"context"
	"fmt"
	"log"

	"github.com/appuio/appuio-cloud-reporting/pkg/testsuite"
)

func main() {
	location, err := testsuite.EnsurePrometheus(context.Background(), testsuite.Version)
	if err != nil {
		log.Panicf("Could not get prometheus: %s", err)
	}
	fmt.Printf("Prometheus exists at path: %s\n", location)
}
