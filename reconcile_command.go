package main

import (
	"fmt"

	"github.com/appuio/appuio-cloud-reporting/pkg/categories"
	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/appuio/appuio-cloud-reporting/pkg/erp"
	"github.com/go-logr/logr"
	"github.com/urfave/cli/v2"
)

type reconcileCommand struct {
	DatabaseURL string
	erpDriver   erp.Driver
}

var reconcileCommandName = "reconcile"

func newReconcileCommand() *cli.Command {
	command := &reconcileCommand{}
	return &cli.Command{
		Name:        reconcileCommandName,
		Usage:       "Perform reconciliations with 3rd party ERP",
		Description: `An ERP adapter has to be loaded as a plugin. Visit the plugin documentation on how to configure it.`,
		Before:      command.before,
		Action:      command.execute,
		After:       command.after,
		Flags: []cli.Flag{
			newDbURLFlag(&command.DatabaseURL),
		},
	}
}

func (cmd *reconcileCommand) before(ctx *cli.Context) error {
	return LogMetadata(ctx)
}

func (cmd *reconcileCommand) execute(context *cli.Context) error {
	log := AppLogger(context.Context).WithName(reconcileCommandName)
	log.V(1).Info("Opening database connection", "url", cmd.DatabaseURL)
	rdb, err := db.Openx(cmd.DatabaseURL)
	if err != nil {
		return fmt.Errorf("could not open database connection: %w", err)
	}

	adapter, err := erp.Get()
	if err != nil {
		return err
	}
	log.V(1).Info("Configuring ERP adapter")
	adapterCtx := logr.NewContext(context.Context, log)
	driver, err := adapter.Open(adapterCtx)
	if err != nil {
		return err
	}
	cmd.erpDriver = driver

	categoryRec := driver.NewCategoryReconciler()
	log.Info("Reconciling categories...")
	return categories.Reconcile(adapterCtx, rdb, categoryRec)
}

func (cmd *reconcileCommand) after(context *cli.Context) error {
	if cmd.erpDriver != nil {
		return cmd.erpDriver.Close(context.Context)
	}
	return nil
}
