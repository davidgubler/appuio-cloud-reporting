package db

import (
	"database/sql"
	"fmt"

	_ "embed"

	"github.com/jmoiron/sqlx"
)

//go:embed seeds/appuio_cloud_memory.promql
var appuioCloudMemoryQuery string

//go:embed seeds/appuio_cloud_loadbalancer.promql
var appuioCloudLoadbalancerQuery string

//go:embed seeds/appuio_cloud_persistent_storage.promql
var appuioCloudPersistentStorageQuery string

//go:embed seeds/ping.promql
var pingQuery string

var DefaultQueries = []Query{
	{
		Name:        "appuio_cloud_memory",
		Description: "Memory usage (maximum of requested and used memory) aggregated by pod",
		Query:       appuioCloudMemoryQuery,
		Unit:        "MiB",
	},
	{
		Name:        "appuio_cloud_loadbalancer",
		Description: "Number of services of type load balancer",
		Query:       appuioCloudLoadbalancerQuery,
	},
	{
		Name:        "appuio_cloud_persistent_storage",
		Description: "Persistent storage usage aggregated by namespace and storageclass",
		Query:       appuioCloudPersistentStorageQuery,
		Unit:        "GiB",
	},
	{
		Name:        "ping",
		Description: "Ping is a query that always returns `42` for `my-product:my-cluster:my-tenant:my-namespace`",
		Query:       pingQuery,
		Unit:        "None",
	},
}

func Seed(db *sql.DB) error {
	dbx := NewDBx(db)
	tx, err := dbx.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := createDefaultQueries(tx); err != nil {
		return err
	}

	if err := createPingQueryDependencies(tx); err != nil {
		return err
	}

	return tx.Commit()
}

func createDefaultQueries(tx *sqlx.Tx) error {
	for _, q := range DefaultQueries {
		exists, err := queryExistsByName(tx, q.Name)
		if err != nil {
			return fmt.Errorf("error checking if query exists: %w", err)
		}
		if exists {
			fmt.Printf("Found query with name '%s'. Skip creating default query.\n", q.Name)
			continue
		}
		_, err = tx.NamedExec("INSERT INTO queries (name,description,query,unit,during) VALUES (:name,:description,:query,:unit,'[-infinity,infinity)')", q)
		if err != nil {
			return fmt.Errorf("error creating default query: %w", err)
		}
	}
	return nil
}

func createPingQueryDependencies(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("INSERT INTO products (source,target,amount,unit,during) VALUES (:source,:target,:amount,:unit,:during) ON CONFLICT DO NOTHING", Product{
		Source: "ping-product:ping-cluster",
		Amount: 1,
		Unit:   "tps",
		During: InfiniteRange(),
	})
	if err != nil {
		return err
	}

	_, err = tx.NamedExec("INSERT INTO discounts (source,discount,during) VALUES (:source,:discount,:during) ON CONFLICT DO NOTHING", Discount{
		Source:   "ping-product:ping-cluster",
		Discount: 0,
		During:   InfiniteRange(),
	})
	if err != nil {
		return err
	}

	return nil
}

func queryExistsByName(tx *sqlx.Tx, name string) (bool, error) {
	var exists bool
	err := tx.Get(&exists, "SELECT EXISTS(SELECT 1 FROM queries WHERE name = $1)", name)
	return exists, err
}
