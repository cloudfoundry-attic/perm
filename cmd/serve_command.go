package cmd

import (
	"crypto/tls"
	"net"

	"strconv"

	"context"

	"time"

	"io/ioutil"
	"os"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/cmd/flags"
	"code.cloudfoundry.org/perm/pkg/api"
	"code.cloudfoundry.org/perm/pkg/api/db"
	"code.cloudfoundry.org/perm/pkg/api/logging"
	"code.cloudfoundry.org/perm/pkg/ioutilx"
	"code.cloudfoundry.org/perm/pkg/sqlx"
)

type ServeCommand struct {
	Logger            flags.LagerFlag
	Hostname          string        `long:"listen-hostname" description:"Hostname on which to listen for gRPC traffic" default:"0.0.0.0"`
	Port              int           `long:"listen-port" description:"Port on which to listen for gRPC traffic" default:"6283"`
	MaxConnectionIdle time.Duration `long:"max-connection-idle" description:"The amount of time before an idle connection will be closed with a GoAway." default:"10s"`
	TLSCertificate    string        `long:"tls-certificate" description:"File path of TLS certificate" required:"true"`
	TLSKey            string        `long:"tls-key" description:"File path of TLS private key" required:"true"`
	SQL               flags.DBFlag  `group:"DB" namespace:"db"`
	AuditFilePath     string        `long:"audit-file-path" default:""`
}

func (cmd ServeCommand) Execute([]string) error {
	//TODO Figure out version dynamically
	version := logging.Version("0.0.0")
	logger, _ := cmd.Logger.Logger("perm")
	logger = logger.Session("serve")

	var auditSink = ioutil.Discard
	if cmd.AuditFilePath != "" {
		securityLogFile, err := ioutilx.OpenLogFile(cmd.AuditFilePath)
		if err != nil {
			return err
		}

		defer securityLogFile.Close()
		auditSink = securityLogFile
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	securityLogger := logging.NewCEFLogger(auditSink, "cloud_foundry", "perm", version, logging.Hostname(hostname), cmd.Port)

	cert, err := tls.LoadX509KeyPair(cmd.TLSCertificate, cmd.TLSKey)
	if err != nil {
		logger.Error(failedToParseTLSCredentials, err)
		return err
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	conn, err := cmd.SQL.Connect(context.Background(), logger)
	if err != nil {
		return err
	}
	defer conn.Close()

	migrationLogger := logger.Session("verify-migrations")
	err = sqlx.VerifyAppliedMigrations(
		context.Background(),
		migrationLogger,
		conn,
		db.MigrationsTableName,
		db.Migrations,
	)
	if err != nil {
		return err
	}

	maxConnectionIdle := cmd.MaxConnectionIdle
	serverOpts := []api.ServerOption{
		api.WithLogger(logger.Session("grpc-serve")),
		api.WithSecurityLogger(securityLogger),
		api.WithTLSConfig(tlsConfig),
		api.WithMaxConnectionIdle(maxConnectionIdle),
	}

	store := db.NewDataService(conn)
	server := api.NewServer(store, serverOpts...)

	listenInterface := cmd.Hostname
	port := cmd.Port
	listeningLogData := lager.Data{
		"protocol":          "tcp",
		"hostname":          listenInterface,
		"port":              port,
		"maxConnectionIdle": maxConnectionIdle.String(),
	}

	lis, err := net.Listen("tcp", net.JoinHostPort(listenInterface, strconv.Itoa(port)))
	if err != nil {
		logger.Error(failedToListen, err, listeningLogData)
		return err
	}

	logger.Debug(starting, listeningLogData)

	return server.Serve(lis)
}
