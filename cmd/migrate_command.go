package cmd

import (
	"context"

	"fmt"
	"os"

	"strconv"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/db"
	"code.cloudfoundry.org/perm/db/migrator"
	"code.cloudfoundry.org/perm/messages"
	"github.com/olekukonko/tablewriter"
)

type MigrateCommand struct {
	Up     UpCommand     `command:"up" description:"Run migrations"`
	Down   DownCommand   `command:"down" description:"Revert migrations"`
	Status StatusCommand `command:"status" description:"Report status of applied migrations"`
}

type UpCommand struct {
	Logger LagerFlag

	SQL SQLFlag `group:"SQL" namespace:"sql"`
}

type DownCommand struct {
	Logger LagerFlag

	SQL SQLFlag `group:"SQL" namespace:"sql"`

	All bool `long:"all" description:"Revert all migrations"`
}

type StatusCommand struct {
	Logger LagerFlag

	SQL SQLFlag `group:"SQL" namespace:"sql"`
}

func (cmd UpCommand) Execute([]string) error {
	logger, _ := cmd.Logger.Logger("perm")
	logger = logger.Session("migrate-up")

	ctx := context.Background()

	conn, err := cmd.SQL.Open()
	if err != nil {
		logger.Error(messages.ErrFailedToOpenSQLConnection, err)
		return err
	}

	pingLogger := logger.Session(messages.PingSQLConnection, cmd.SQL.LagerData())
	pingLogger.Debug(messages.Starting)
	err = conn.PingContext(ctx)
	if err != nil {
		pingLogger.Error(messages.ErrFailedToPingSQLConnection, err, cmd.SQL.LagerData())
		return err
	}
	pingLogger.Debug(messages.Finished)

	defer conn.Close()

	return migrator.ApplyMigrations(ctx, logger, conn, db.MigrationsTableName, db.Migrations)
}

func (cmd DownCommand) Execute([]string) error {
	logger, _ := cmd.Logger.Logger("perm")
	logger = logger.Session("migrate-down").WithData(lager.Data{
		"all": cmd.All,
	})

	ctx := context.Background()

	conn, err := cmd.SQL.Open()
	if err != nil {
		logger.Error(messages.ErrFailedToOpenSQLConnection, err)
		return err
	}

	pingLogger := logger.Session(messages.PingSQLConnection, cmd.SQL.LagerData())
	pingLogger.Debug(messages.Starting)
	err = conn.PingContext(ctx)
	if err != nil {
		pingLogger.Error(messages.ErrFailedToPingSQLConnection, err, cmd.SQL.LagerData())
		return err
	}
	pingLogger.Debug(messages.Finished)

	defer conn.Close()

	return migrator.RollbackMigrations(ctx, logger, conn, db.MigrationsTableName, db.Migrations, cmd.All)
}

func (cmd StatusCommand) Execute([]string) error {
	logger, _ := cmd.Logger.Logger("perm")
	logger = logger.Session("migrate-status")

	ctx := context.Background()

	conn, err := cmd.SQL.Open()
	if err != nil {
		logger.Error(messages.ErrFailedToOpenSQLConnection, err)
		return err
	}

	pingLogger := logger.Session(messages.PingSQLConnection, cmd.SQL.LagerData())
	pingLogger.Debug(messages.Starting)
	err = conn.PingContext(ctx)
	if err != nil {
		pingLogger.Error(messages.ErrFailedToPingSQLConnection, err, cmd.SQL.LagerData())
		return err
	}
	pingLogger.Debug(messages.Finished)

	defer conn.Close()

	appliedMigrations, err := migrator.RetrieveAppliedMigrations(ctx, logger, conn, db.MigrationsTableName)
	if err != nil {
		return err
	}

	f := os.Stdout

	fmt.Fprintln(f, "Applied Migrations")
	appliedMigrationsTable := tablewriter.NewWriter(f)
	appliedMigrationsTable.SetHeader([]string{"Version", "Name", "Applied At"})
	for i, migration := range db.Migrations {
		version := i

		appliedMigration, ok := appliedMigrations[version]
		if ok {
			appliedMigrationsTable.Append([]string{strconv.Itoa(version), migration.Name, appliedMigration.AppliedAt.Local().String()})
		}
	}
	appliedMigrationsTable.Render()

	fmt.Fprintln(f, "\nMigrations Not Yet Applied")
	unappliedMigrationsTable := tablewriter.NewWriter(f)
	unappliedMigrationsTable.SetHeader([]string{"Version", "Name"})
	for i, migration := range db.Migrations {
		version := i

		_, ok := appliedMigrations[version]
		if !ok {
			unappliedMigrationsTable.Append([]string{strconv.Itoa(version), migration.Name})
		}
	}
	unappliedMigrationsTable.Render()

	return nil
}
