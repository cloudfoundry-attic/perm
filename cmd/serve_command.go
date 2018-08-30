package cmd

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"strconv"

	"context"

	"time"

	"io/ioutil"
	"os"

	"code.cloudfoundry.org/perm/cmd/flags"
	"code.cloudfoundry.org/perm/internal/migrations"
	"code.cloudfoundry.org/perm/internal/sqlx"
	"code.cloudfoundry.org/perm/pkg/api"
	"code.cloudfoundry.org/perm/pkg/cryptox"
	"code.cloudfoundry.org/perm/pkg/ioutilx"
	"code.cloudfoundry.org/perm/pkg/logx"
	"code.cloudfoundry.org/perm/pkg/logx/cef"
	"code.cloudfoundry.org/perm/pkg/metrics/statsdx"
	"code.cloudfoundry.org/perm/pkg/permauth"
	"github.com/cactus/go-statsd-client/statsd"
	oidc "github.com/coreos/go-oidc"
)

type ServeCommand struct {
	Logger            flags.LagerFlag
	StatsD            statsDOptions        `group:"StatsD" namespace:"statsd"`
	Hostname          string               `long:"listen-hostname" description:"Hostname on which to listen for gRPC traffic" default:"0.0.0.0"`
	Port              int                  `long:"listen-port" description:"Port on which to listen for gRPC traffic" default:"6283"`
	MaxConnectionIdle time.Duration        `long:"max-connection-idle" description:"The amount of time before an idle connection will be closed with a GoAway." default:"10s"`
	TLSCertificate    string               `long:"tls-certificate" description:"File path of TLS certificate" required:"true"`
	TLSKey            string               `long:"tls-key" description:"File path of TLS private key" required:"true"`
	DB                flags.DBFlag         `group:"DB" namespace:"db"`
	AuditFilePath     string               `long:"audit-file-path" default:""`
	OAuth2URL         string               `long:"oauth2-url" description:"URL of the OAuth2 provider (only required if '--required-auth' is provided)"`
	OAuth2CA          ioutilx.FileOrString `long:"oauth2-certificate-authority" description:"the certificate authority of the OAuth2 provider (only required if '--required-auth' is provided)"`
	RequireAuth       bool                 `long:"require-auth" description:"Require auth"`
}

type statsDOptions struct {
	Hostname string `long:"hostname" description:"Hostname used to connect to StatsD server"`
	Port     int    `long:"port" description:"Port used to connect to StatsD server" default:"8125"`
}

func (cmd ServeCommand) Execute([]string) error {
	//TODO Figure out version dynamically
	version := cef.Version("0.0.0")
	logger := cmd.Logger.Logger("perm")

	var auditSink = ioutil.Discard
	if cmd.AuditFilePath != "" {
		securityLogFile, err := ioutilx.OpenLogFile(cmd.AuditFilePath)
		if err != nil {
			return err
		}

		defer securityLogFile.Close()
		auditSink = securityLogFile
	}

	hostname, hostErr := os.Hostname()
	if hostErr != nil {
		return hostErr
	}

	securityLogger := cef.NewLogger(auditSink, "cloud_foundry", "perm", version, cef.Hostname(hostname), cmd.Port, logger.WithName("cef"))

	cert, certErr := tls.LoadX509KeyPair(cmd.TLSCertificate, cmd.TLSKey)
	if certErr != nil {
		logger.Error(failedToParseTLSCredentials, certErr)
		return certErr
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	serverOpts := []api.ServerOption{
		api.WithLogger(logger),
		api.WithSecurityLogger(securityLogger),
		api.WithTLSConfig(tlsConfig),
		api.WithMaxConnectionIdle(cmd.MaxConnectionIdle),
	}

	if cmd.StatsD.Hostname != "" {
		statsDAddr := net.JoinHostPort(cmd.StatsD.Hostname, strconv.Itoa(cmd.StatsD.Port))
		statsDClient, err := statsd.NewClient(statsDAddr, "")
		if err != nil {
			logger.Error("failed-to-connect-to-statsd", err, logx.Data{
				Key:   "address",
				Value: statsDAddr,
			})
		} else {
			defer statsDClient.Close()

			statter := statsdx.NewStatter(logger.WithName("statsd"), statsDClient)
			serverOpts = append(serverOpts, api.WithStatter(statter))
		}
	}

	if !cmd.DB.IsInMemory() {
		conn, err := cmd.DB.Connect(context.Background(), logger)
		if err != nil {
			return err
		}
		defer conn.Close()

		migrationLogger := logger.WithName("verify-migrations")
		if err := sqlx.VerifyAppliedMigrations(
			context.Background(),
			migrationLogger,
			conn,
			migrations.TableName,
			migrations.Migrations,
		); err != nil {
			return err
		}

		serverOpts = append(serverOpts, api.WithDBConn(conn))
	}

	if cmd.RequireAuth {
		oauthCA, err := cmd.OAuth2CA.Bytes(ioutilx.OS, ioutilx.IOReader)
		if err != nil {
			return err
		}

		oauthCAPool, err := cryptox.NewCertPool(oauthCA)
		if err != nil {
			return err
		}

		tlsClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: oauthCAPool,
				},
			},
		}
		oidcContext := oidc.ClientContext(context.Background(), tlsClient)
		oidcIssuer, err := permauth.GetOIDCIssuer(tlsClient, fmt.Sprintf("%s/oauth/token", cmd.OAuth2URL))
		if err != nil {
			return err
		}
		provider, err := oidc.NewProvider(oidcContext, oidcIssuer)
		if err != nil {
			return err
		}
		serverOpts = append(serverOpts, api.WithOIDCProvider(provider))
	}

	server := api.NewServer(serverOpts...)

	listenInterface := cmd.Hostname
	port := cmd.Port
	listeningLogData := []logx.Data{
		logx.Data{Key: "protocol", Value: "tcp"},
		logx.Data{Key: "hostname", Value: listenInterface},
		logx.Data{Key: "port", Value: port},
		logx.Data{Key: "MaxConnectionIdle", Value: cmd.MaxConnectionIdle},
	}

	lis, err := net.Listen("tcp", net.JoinHostPort(listenInterface, strconv.Itoa(port)))
	if err != nil {
		logger.Error(failedToListen, err, listeningLogData...)
		return err
	}

	logger.Debug(starting, listeningLogData...)

	return server.Serve(lis)
}
