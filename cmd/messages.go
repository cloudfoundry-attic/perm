package cmd

const (
	errInternal = "internal"

	starting = "starting"
	finished = "finished"

	failedToListen              = "failed-to-listen"
	failedToParseTLSCredentials = "failed-to-parse-tls-credentials"

	openSQLConnection         = "open-sql-connection"
	failedToOpenSQLConnection = "failed-to-open-sql-connection"
	pingSQLConnection         = "ping-sql-connection"
	failedToPingSQLConnection = "failed-to-ping-sql-connection"
)
