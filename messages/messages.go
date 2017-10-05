package messages

const (
	ErrFailedToListen              = "failed-to-listen"
	ErrFailedToParseOptions        = "failed-to-parse-options"
	ErrInternal                    = "internal"
	ErrInvalidTLSCredentials       = "invalid-tls-credentials"
	ErrRoleAlreadyExists           = "role-already-exists"
	ErrRoleAssignmentAlreadyExists = "role-assignment-already-exists"
	ErrRoleAssignmentNotFound      = "role-assignment-not-found"
	ErrRoleNotFound                = "role-not-found"

	StartingServer = "starting-server"
	Success        = "success"
)

const (
	FailedToConnectToStatsD = "failed-to-connect-to-statsd"
	FailedToSendMetric      = "failed-to-send-metric"

	FailedToReadCertificate  = "failed-to-read-certificate"
	FailedToAppendCertToPool = "failed-to-append-cert-to-pool"
	FailedToGRPCDial         = "failed-to-grpc-dial"

	FailedToCreateRole   = "failed-to-create-role"
	FailedToAssignRole   = "failed-to-assign-role"
	FailedToUnassignRole = "failed-to-unassign-role"
	FailedToDeleteRole   = "failed-to-delete-role"
)
