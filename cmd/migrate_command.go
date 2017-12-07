package cmd

import (
	"context"

	"fmt"
	"os"

	"strconv"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/db"
	"code.cloudfoundry.org/perm/sqlx"
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

	conn, err := cmd.SQL.Connect(ctx, logger, OS, IOReader)
	if err != nil {
		return err
	}
	defer conn.Close()

	return sqlx.ApplyMigrations(ctx, logger, conn, db.MigrationsTableName, db.Migrations)
}

func (cmd DownCommand) Execute([]string) error {
	logger, _ := cmd.Logger.Logger("perm")
	logger = logger.Session("migrate-down").WithData(lager.Data{
		"all": cmd.All,
	})

	ctx := context.Background()

	conn, err := cmd.SQL.Connect(ctx, logger, OS, IOReader)
	if err != nil {
		return err
	}
	defer conn.Close()

	return sqlx.RollbackMigrations(ctx, logger, conn, db.MigrationsTableName, db.Migrations, cmd.All)
}

func (cmd StatusCommand) Execute([]string) error {
	logger, _ := cmd.Logger.Logger("perm")
	logger = logger.Session("migrate-status")

	ctx := context.Background()

	conn, err := cmd.SQL.Connect(ctx, logger, OS, IOReader)
	if err != nil {
		return err
	}
	defer conn.Close()

	appliedMigrations, err := sqlx.RetrieveAppliedMigrations(ctx, logger, conn, db.MigrationsTableName)
	if err != nil {
		return err
	}

	f := os.Stdout

	appliedMigrationsTable := tablewriter.NewWriter(f)
	appliedMigrationsTable.SetHeader([]string{"Version", "Name", "Applied At"})

	unappliedMigrationsTable := tablewriter.NewWriter(f)
	unappliedMigrationsTable.SetHeader([]string{"Version", "Name"})
	for i, migration := range db.Migrations {
		version := i

		appliedMigration, ok := appliedMigrations[version]
		if ok {
			appliedMigrationsTable.Append([]string{strconv.Itoa(version), migration.Name, appliedMigration.AppliedAt.Local().String()})
		} else {
			unappliedMigrationsTable.Append([]string{strconv.Itoa(version), migration.Name})
		}
	}

	fmt.Fprintln(f, "Applied Migrations")
	appliedMigrationsTable.Render()

	fmt.Println("")
	fmt.Fprintln(f, "Migrations Not Yet Applied")
	unappliedMigrationsTable.Render()

	return nil
}
