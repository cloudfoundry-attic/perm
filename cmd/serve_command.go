package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/cmd/flags"
	"code.cloudfoundry.org/perm/pkg/api"
	"code.cloudfoundry.org/perm/pkg/api/db"
	"code.cloudfoundry.org/perm/pkg/api/logging"
	"code.cloudfoundry.org/perm/pkg/cryptox"
	"code.cloudfoundry.org/perm/pkg/ioutilx"
	"code.cloudfoundry.org/perm/pkg/sqlx"
	oidc "github.com/coreos/go-oidc"
)

type ServeCommand struct {
	Logger            flags.LagerFlag
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

	maxConnectionIdle := cmd.MaxConnectionIdle
	serverOpts := []api.ServerOption{
		api.WithLogger(logger.Session("grpc-serve")),
		api.WithSecurityLogger(securityLogger),
		api.WithTLSConfig(tlsConfig),
		api.WithMaxConnectionIdle(maxConnectionIdle),
	}

	if !cmd.DB.IsInMemory() {
		conn, err := cmd.DB.Connect(context.Background(), logger)
		if err != nil {
			return err
		}
		defer conn.Close()

		migrationLogger := logger.Session("verify-migrations")
		if err := sqlx.VerifyAppliedMigrations(
			context.Background(),
			migrationLogger,
			conn,
			db.MigrationsTableName,
			db.Migrations,
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

		oidcContext := oidc.ClientContext(context.Background(), &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: oauthCAPool,
				},
			},
		})

		oauth2URL, err := removeSchemeSpecificPort(cmd.OAuth2URL)
		if err != nil {
			return err
		}

		provider, err := oidc.NewProvider(oidcContext, fmt.Sprintf("%s/oauth/token", oauth2URL))
		if err != nil {
			return err
		}
		serverOpts = append(serverOpts, api.WithOIDCProvider(provider))
	}

	server := api.NewServer(serverOpts...)

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

func removeSchemeSpecificPort(urlString string) (string, error) {
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return "", err
	}

	parts := strings.Split(parsedURL.Host, ":")
	if len(parts) != 2 {
		return urlString, nil
	}

	if (parsedURL.Scheme == "http" && parts[1] == "80") ||
		(parsedURL.Scheme == "https" && parts[1] == "443") {
		parsedURL.Host = parts[0]
	}

	return parsedURL.String(), nil
}
