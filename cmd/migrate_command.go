package cmd

import (
	"context"

	"fmt"
	"os"

	"strconv"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/cmd/flags"
	"code.cloudfoundry.org/perm/pkg/api/db"
	"code.cloudfoundry.org/perm/pkg/sqlx"
	"github.com/olekukonko/tablewriter"
)

type MigrateCommand struct {
	Up     UpCommand     `command:"up" description:"Run migrations"`
	Down   DownCommand   `command:"down" description:"Revert migrations"`
	Status StatusCommand `command:"status" description:"Report status of applied migrations"`
}

type UpCommand struct {
	Logger flags.LagerFlag

	DB flags.DBFlag `group:"DB" namespace:"db"`
}

type DownCommand struct {
	Logger flags.LagerFlag

	DB flags.DBFlag `group:"DB" namespace:"db"`

	All bool `long:"all" description:"Revert all migrations"`
}

type StatusCommand struct {
	Logger flags.LagerFlag

	DB flags.DBFlag `group:"DB" namespace:"db"`
}

func (cmd UpCommand) Execute([]string) error {
	logger, _ := cmd.Logger.Logger("perm")
	logger = logger.Session("migrate-up")

	if cmd.DB.IsInMemory() {
		return nil
	}

	ctx := context.Background()

	conn, err := cmd.DB.Connect(ctx, logger)
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

	if cmd.DB.IsInMemory() {
		return nil
	}

	ctx := context.Background()

	conn, err := cmd.DB.Connect(ctx, logger)
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
	conn, err := cmd.DB.Connect(ctx, logger)
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

		migrationVersion := strconv.Itoa(version)
		migrationName := migration.Name

		if ok {
			appliedAtTime := appliedMigration.AppliedAt.Local().String()
			appliedMigrationsTable.Append([]string{migrationVersion, migrationName, appliedAtTime})
		} else {
			unappliedMigrationsTable.Append([]string{migrationVersion, migrationName})
		}
	}

	fmt.Fprintln(f, "Applied Migrations")
	appliedMigrationsTable.Render()

	fmt.Println("")
	fmt.Fprintln(f, "Migrations Not Yet Applied")
	unappliedMigrationsTable.Render()

	return nil
}
