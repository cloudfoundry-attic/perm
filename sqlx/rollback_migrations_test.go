package sqlx_test

import (
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	. "code.cloudfoundry.org/perm/sqlx"

	"context"
	"database/sql"

	"errors"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("#RollbackMigrations", func() {
	var (
		migrationTableName string

		fakeLogger *lagertest.TestLogger

		fakeConn *sql.DB
		mock     sqlmock.Sqlmock
		err      error

		conn *DB

		ctx context.Context

		migrations []Migration

		all bool

		appliedAt time.Time
	)

	BeforeEach(func() {
		migrationTableName = "db_migrations"

		fakeLogger = lagertest.NewTestLogger("perm-sqlx")

		fakeConn, mock, err = sqlmock.New()
		Expect(err).NotTo(HaveOccurred())

		conn = &DB{
			DB: fakeConn,
		}

		appliedAt = time.Now()

		ctx = context.Background()

		migrations = []Migration{
			{
				Name: "migration_1",
				Down: func(ctx context.Context, logger lager.Logger, tx *Tx) error {
					_, err := tx.ExecContext(ctx, "SOME FAKE MIGRATION 1")

					return err
				},
			},
			{
				Name: "migration_2",
				Down: func(ctx context.Context, logger lager.Logger, tx *Tx) error {
					_, err := tx.ExecContext(ctx, "SOME FAKE MIGRATION 2")

					return err
				},
			},
			{
				Name: "migration_3",
				Down: func(ctx context.Context, logger lager.Logger, tx *Tx) error {
					_, err := tx.ExecContext(ctx, "SOME FAKE MIGRATION 3")

					return err
				},
			},
		}
	})

	AfterEach(func() {
		Expect(mock.ExpectationsWereMet()).To(Succeed())
	})

	Context("without 'all'", func() {

		BeforeEach(func() {
			all = false
		})

		It("rolls back the most recent migration which is in the migrations table", func() {
			mock.ExpectQuery("SELECT version, name, applied_at FROM " + migrationTableName).
				WillReturnRows(
					sqlmock.NewRows([]string{"version", "name", "applied_at"}).
						AddRow("0", "migration_1", appliedAt).
						AddRow("1", "migration_2", appliedAt),
				)

			mock.ExpectBegin()
			mock.ExpectExec("SOME FAKE MIGRATION 2").
				WillReturnResult(sqlmock.NewResult(0, 0))
			mock.ExpectExec("DELETE FROM " + migrationTableName + " WHERE version").
				WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 0))
			mock.ExpectCommit()

			err := RollbackMigrations(ctx, fakeLogger, conn, migrationTableName, migrations, all)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("with 'all'", func() {
		var (
			all bool
		)

		BeforeEach(func() {
			all = true
		})

		It("rolls back all migrations found in the database", func() {
			mock.ExpectQuery("SELECT version, name, applied_at FROM " + migrationTableName).
				WillReturnRows(
					sqlmock.NewRows([]string{"version", "name", "applied_at"}).
						AddRow("0", "migration_1", appliedAt).
						AddRow("1", "migration_2", appliedAt),
				)

			mock.ExpectBegin()
			mock.ExpectExec("SOME FAKE MIGRATION 2").
				WillReturnResult(sqlmock.NewResult(0, 0))
			mock.ExpectExec("DELETE FROM " + migrationTableName + " WHERE version").
				WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 0))
			mock.ExpectCommit()
			mock.ExpectBegin()
			mock.ExpectExec("SOME FAKE MIGRATION 1").
				WillReturnResult(sqlmock.NewResult(0, 0))
			mock.ExpectExec("DELETE FROM " + migrationTableName + " WHERE version").
				WithArgs(0).WillReturnResult(sqlmock.NewResult(0, 0))
			mock.ExpectCommit()

			err := RollbackMigrations(ctx, fakeLogger, conn, migrationTableName, migrations, all)
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not run earlier migrations when a migration fails", func() {
			mock.ExpectQuery("SELECT version, name, applied_at FROM " + migrationTableName).
				WillReturnRows(
					sqlmock.NewRows([]string{"version", "name", "applied_at"}).
						AddRow("0", "migration_1", appliedAt).
						AddRow("1", "migration_2", appliedAt),
				)

			mock.ExpectBegin()
			mock.ExpectExec("SOME FAKE MIGRATION 2").WillReturnError(errors.New("some-rollback-error"))
			mock.ExpectRollback()

			err := RollbackMigrations(ctx, fakeLogger, conn, migrationTableName, migrations, all)
			Expect(err).To(MatchError("some-rollback-error"))
		})

		It("does not run earlier migrations when a commit fails", func() {
			mock.ExpectQuery("SELECT version, name, applied_at FROM " + migrationTableName).
				WillReturnRows(
					sqlmock.NewRows([]string{"version", "name", "applied_at"}).
						AddRow("0", "migration_1", appliedAt).
						AddRow("1", "migration_2", appliedAt),
				)

			mock.ExpectBegin()
			mock.ExpectExec("SOME FAKE MIGRATION 2").
				WillReturnResult(sqlmock.NewResult(0, 0))
			mock.ExpectExec("DELETE FROM " + migrationTableName + " WHERE version").
				WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 0))
			mock.ExpectCommit().WillReturnError(errors.New("some-commit-error"))

			err := RollbackMigrations(ctx, fakeLogger, conn, migrationTableName, migrations, all)
			Expect(err).To(MatchError("some-commit-error"))
		})
	})
})
