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

var _ = Describe("#VerifyAppliedMigrations", func() {
	var (
		migrationTableName string

		fakeLogger *lagertest.TestLogger

		fakeConn *sql.DB
		mock     sqlmock.Sqlmock
		err      error

		ctx context.Context

		migrations []Migration

		appliedAt time.Time
	)

	BeforeEach(func() {
		migrationTableName = "db_migrations"

		fakeLogger = lagertest.NewTestLogger("perm-sqlx")

		fakeConn, mock, err = sqlmock.New()
		Expect(err).NotTo(HaveOccurred())

		appliedAt = time.Now()

		ctx = context.Background()

		migrations = []Migration{
			{
				Name: "migration_1",
				Down: func(ctx context.Context, logger lager.Logger, tx *sql.Tx) error {
					_, err := tx.ExecContext(ctx, "SOME FAKE MIGRATION 1")

					return err
				},
			},
			{
				Name: "migration_2",
				Down: func(ctx context.Context, logger lager.Logger, tx *sql.Tx) error {
					_, err := tx.ExecContext(ctx, "SOME FAKE MIGRATION 2")

					return err
				},
			},
			{
				Name: "migration_3",
				Down: func(ctx context.Context, logger lager.Logger, tx *sql.Tx) error {
					_, err := tx.ExecContext(ctx, "SOME FAKE MIGRATION 3")

					return err
				},
			},
		}
	})

	AfterEach(func() {
		Expect(mock.ExpectationsWereMet()).To(Succeed())
	})

	It("returns true if there are no migrations", func() {
		mock.ExpectQuery("SELECT version, name, applied_at FROM " + migrationTableName).
			WillReturnRows(sqlmock.NewRows([]string{"version", "name", "applied_at"}))

		verify, err := VerifyAppliedMigrations(ctx, fakeLogger, fakeConn, migrationTableName, []Migration{})

		Expect(err).NotTo(HaveOccurred())
		Expect(verify).To(BeTrue())
	})

	It("returns true if all the migrations match", func() {
		mock.ExpectQuery("SELECT version, name, applied_at FROM " + migrationTableName).
			WillReturnRows(sqlmock.NewRows([]string{"version", "name", "applied_at"}).
				AddRow("0", "migration_1", appliedAt).
				AddRow("1", "migration_2", appliedAt).
				AddRow("2", "migration_3", appliedAt),
			)

		verify, err := VerifyAppliedMigrations(ctx, fakeLogger, fakeConn, migrationTableName, migrations)

		Expect(err).NotTo(HaveOccurred())
		Expect(verify).To(BeTrue())
	})

	It("returns false if there is a migration count mismatch", func() {
		mock.ExpectQuery("SELECT version, name, applied_at FROM " + migrationTableName).
			WillReturnRows(sqlmock.NewRows([]string{"version", "name", "applied_at"}).
				AddRow("0", "migration_1", appliedAt).
				AddRow("1", "migration_2", appliedAt),
			)

		verify, err := VerifyAppliedMigrations(ctx, fakeLogger, fakeConn, migrationTableName, migrations)

		Expect(err).NotTo(HaveOccurred())
		Expect(verify).To(BeFalse())
	})

	It("returns false if there is a migration which has not been applied", func() {
		mock.ExpectQuery("SELECT version, name, applied_at FROM " + migrationTableName).
			WillReturnRows(sqlmock.NewRows([]string{"version", "name", "applied_at"}).
				AddRow("0", "migration_1", appliedAt).
				AddRow("2", "migration_2", appliedAt).
				AddRow("3", "migration_3", appliedAt),
			)

		verify, err := VerifyAppliedMigrations(ctx, fakeLogger, fakeConn, migrationTableName, migrations)

		Expect(err).NotTo(HaveOccurred())
		Expect(verify).To(BeFalse())
	})

	It("returns false if the migration names do not match in order of application", func() {
		mock.ExpectQuery("SELECT version, name, applied_at FROM " + migrationTableName).
			WillReturnRows(sqlmock.NewRows([]string{"version", "name", "applied_at"}).
				AddRow("0", "migration_2", appliedAt).
				AddRow("1", "migration_1", appliedAt).
				AddRow("2", "migration_3", appliedAt),
			)

		verify, err := VerifyAppliedMigrations(ctx, fakeLogger, fakeConn, migrationTableName, migrations)

		Expect(err).NotTo(HaveOccurred())
		Expect(verify).To(BeFalse())
	})

	It("fails if it cannot retrieve the applied migrations", func() {
		mock.ExpectQuery("SELECT version, name, applied_at FROM " + migrationTableName).
			WillReturnError(errors.New("some sql error"))

		_, err := VerifyAppliedMigrations(ctx, fakeLogger, fakeConn, migrationTableName, migrations)

		Expect(err).To(MatchError("some sql error"))
	})
})
