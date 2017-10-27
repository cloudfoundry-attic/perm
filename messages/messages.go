package messages

// `Err` messages are for states that are expected as part of the normal control flow,
// e.g., a query returned no results.
// `Failed` messages are for actions that were expected to succeed but erred instead,
// e.g., the database connection failed.

const (
	ErrInternal = "internal"

	Starting = "starting"
	Finished = "finished"
	Success  = "success"
)

// External integrations messages
const (
	FailedToListen              = "failed-to-listen"
	FailedToParseTLSCredentials = "failed-to-parse-tls-credentials"

	PingSQLConnection = "ping-sql-connection"

	FailedToOpenSQLConnection = "failed-to-open-sql-connection"
	FailedToPingSQLConnection = "failed-to-ping-sql-connection"

	FailedToConnectToStatsD = "failed-to-connect-to-statsd"
	FailedToSendMetric      = "failed-to-send-metric"

	FailedToReadCertificate  = "failed-to-read-certificate"
	FailedToAppendCertToPool = "failed-to-append-cert-to-pool"
	FailedToGRPCDial         = "failed-to-grpc-dial"
)

// Resource messages
const (
	ErrRoleAlreadyExists = "role-already-exists"
	ErrRoleNotFound      = "role-not-found"

	FailedToCreateRole   = "failed-to-create-role"
	FailedToFindRole     = "failed-to-find-role"
	FailedToDeleteRole   = "failed-to-delete-role"
	FailedToAssignRole   = "failed-to-assign-role"
	FailedToUnassignRole = "failed-to-unassign-role"

	ErrActorAlreadyExists = "actor-already-exists"
	ErrActorNotFound      = "actor-not-found"

	FailedToCreateActor = "failed-to-create-actor"
	FailedToFindActor   = "failed-to-find-actor"

	ErrRoleAssignmentAlreadyExists = "role-assignment-already-exists"
	ErrRoleAssignmentNotFound      = "role-assignment-not-found"

	FailedToCreateRoleAssignment = "failed-to-create-role-assignment"
	FailedToDeleteRoleAssignment = "failed-to-delete-role-assignment"
	FailedToFindRoleAssignment   = "failed-to-find-role-assignment"
)

// Migration messages
const (
	FailedToApplyMigration  = "failed-to-apply-migration"
	FailedToQueryMigrations = "failed-to-query-migrations"
	FailedToCreateTable     = "failed-to-create-table"

	FailedToParseAppliedMigration = "failed-to-parse-applied-migration"

	RetrievedAppliedMigrations = "retrieved-applied-migrations"
	SkippedAppliedMigration    = "skipped-applied-migration"
)

// Database messages
const (
	Committed = "committed"

	FailedToStartTransaction = "failed-to-start-transaction"
	FailedToCommit           = "failed-to-commit"
	FailedToRollback         = "failed-to-rollback"

	FailedToRetrieveID        = "failed-to-retrieve-id"
	FailedToCountRowsAffected = "failed-to-count-rows-affected"
)
