package messages

const (
	ErrFailedToListen            = "failed-to-listen"
	ErrFailedToOpenSQLConnection = "failed-to-open-sql-connection"
	ErrFailedToPingSQLConnection = "failed-to-ping-sql-connection"
	ErrInternal                  = "internal"
	ErrInvalidTLSCredentials     = "invalid-tls-credentials"
)

const (
	Starting = "starting"
	Finished = "finished"
	Success  = "success"

	FailedToConnectToStatsD = "failed-to-connect-to-statsd"
	FailedToSendMetric      = "failed-to-send-metric"

	FailedToReadCertificate  = "failed-to-read-certificate"
	FailedToAppendCertToPool = "failed-to-append-cert-to-pool"
	FailedToGRPCDial         = "failed-to-grpc-dial"

	ErrRoleAlreadyExists = "role-already-exists"
	ErrRoleNotFound      = "role-not-found"

	FailedToCreateRole   = "failed-to-create-role"
	FailedToAssignRole   = "failed-to-assign-role"
	FailedToUnassignRole = "failed-to-unassign-role"
	FailedToDeleteRole   = "failed-to-delete-role"
	FailedToFindRole     = "failed-to-find-role"

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

const (
	PingSQLConnection = "ping-sql-connection"
)

const (
	ErrFailedToApplyMigration  = "failed-to-run-migration"
	ErrFailedToQueryMigrations = "failed-to-query-migrations"
	ErrFailedToCreateTable     = "failed-to-create-table"

	ErrFailedToParseAppliedMigration = "failed-to-parse-applied-migration"

	RetrievedAppliedMigrations = "retrieved-applied-migrations"
	SkippedAppliedMigration    = "skipped-applied-migration"
)

const (
	Committed = "committed"

	FailedToStartTransaction = "failed-to-start-transaction"
	FailedToCommit           = "failed-to-commit"
	FailedToRollback         = "failed-to-rollback"

	FailedToRetrieveID        = "failed-to-retrieve-id"
	FailedToCountRowsAffected = "failed-to-count-rows-affected"
)
