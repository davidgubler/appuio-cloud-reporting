package promflag

import (
	"flag"
	"os"

	"github.com/prometheus/client_golang/api"
	apiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

var PrometheusURL string

func init() {
	flag.StringVar(&PrometheusURL, "prom-url", urlFromEnv(), "The URL to connect to prometheus.")
}

func urlFromEnv() string {
	if url, exists := os.LookupEnv("PROM_URL"); exists {
		return url
	}
	return "http://127.0.0.1:9090"
}

func PrometheusAPIClient() (apiv1.API, error) {
	client, err := api.NewClient(api.Config{
		Address: PrometheusURL,
	})
	if err != nil {
		return nil, err
	}
	return apiv1.NewAPI(client), nil
}
