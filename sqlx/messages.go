package sqlx

const (
	starting = "starting"
	finished = "finished"
	success  = "success"

	committed                = "committed"
	failedToStartTransaction = "failed-to-start-transaction"
	failedToCommit           = "failed-to-commit"
	failedToRollback         = "failed-to-rollback"

	retrievedAppliedMigrations    = "retrieved-applied-migrations"
	skippedAppliedMigration       = "skipped-applied-migration"
	failedToApplyMigration        = "failed-to-apply-migration"
	failedToQueryMigrations       = "failed-to-query-migrations"
	failedToParseAppliedMigration = "failed-to-parse-applied-migration"

	failedToCreateTable = "failed-to-create-table"

	migrationCountMismatch = "migration-count-mismatch"
	migrationNotFound      = "migration-not-found"
	migrationMismatch      = "migration-mismatch"
)
