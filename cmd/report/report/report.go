package report

import (
	"flag"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	apiv1 "github.com/prometheus/client_golang/api/prometheus/v1"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	dbflag "github.com/appuio/appuio-cloud-reporting/pkg/db/flag"
	"github.com/appuio/appuio-cloud-reporting/pkg/report"
)

func Main() error {
	flag.Parse()

	client, err := api.NewClient(api.Config{
		Address: "http://demo.robustperception.io:9090",
	})
	clientv1 := apiv1.NewAPI(client)
	if err != nil {
		return fmt.Errorf("could not create prometheus client: %w", err)
	}

	rdb, err := db.Openx(dbflag.DatabaseURL)
	if err != nil {
		return fmt.Errorf("could not open database connection: %w", err)
	}
	defer rdb.Close()

	return report.Run(rdb, clientv1, "test", time.Now())
}
