package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"os"

	"code.cloudfoundry.org/perm/cmd"
	"code.cloudfoundry.org/perm/messages"
	"code.cloudfoundry.org/perm/protos"
	"code.cloudfoundry.org/perm/rpc"

	"code.cloudfoundry.org/lager"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/jessevdk/go-flags"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	"strconv"

	"github.com/go-sql-driver/mysql"
)

type options struct {
	Hostname       string `long:"listen-hostname" description:"Hostname on which to listen for gRPC traffic" default:"0.0.0.0"`
	Port           int    `long:"listen-port" description:"Port on which to listen for gRPC traffic" default:"6283"`
	TLSCertificate string `long:"tls-certificate" description:"File path of TLS certificate" required:"true"`
	TLSKey         string `long:"tls-key" description:"File path of TLS private key" required:"true"`
	Logger         cmd.LagerFlag
	SQL            sqlOptions `group:"SQL" namespace:"sql"`
}

type sqlOptions struct {
	DB dbOptions `group:"DB" namespace:"db"`
}

type dbOptions struct {
	Driver   string `long:"driver" description:"Database driver to use for SQL backend (e.g. mysql, postgres)" required:"true"`
	Host     string `long:"host" description:"Host for SQL backend" required:"true"`
	Port     int    `long:"port" description:"Port for SQL backend" required:"true"`
	Schema   string `long:"schema" description:"Database name to use for connecting to SQL backend" required:"true"`
	Username string `long:"username" description:"Username to use for connecting to SQL backend" required:"true"`
	Password string `long:"password" description:"Password to use for connecting to SQL backend" required:"true"`
}

func main() {
	parserOpts := &options{}
	parser := flags.NewParser(parserOpts, flags.Default)
	parser.NamespaceDelimiter = "-"

	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}

	logger, _ := parserOpts.Logger.Logger("perm")

	hostname := parserOpts.Hostname
	port := parserOpts.Port
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", hostname, port))

	listeningLogData := lager.Data{
		"protocol": "tcp",
		"hostname": hostname,
		"port":     port,
	}
	if err != nil {
		logger.Fatal(messages.ErrFailedToListen, err, listeningLogData)
		os.Exit(1)
	}

	tlsCreds, err := credentials.NewServerTLSFromFile(parserOpts.TLSCertificate, parserOpts.TLSKey)

	if err != nil {
		logger.Fatal(messages.ErrInvalidTLSCredentials, err)
		os.Exit(1)
	}

	recoveryOpts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(func(p interface{}) error {
			err := status.Errorf(codes.Internal, "%s", p)
			logger.Error(messages.ErrInternal, err)
			return err
		}),
	}
	serverOpts := []grpc.ServerOption{
		grpc.Creds(tlsCreds),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(grpc_recovery.StreamServerInterceptor(recoveryOpts...))),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(grpc_recovery.UnaryServerInterceptor(recoveryOpts...))),
	}

	grpcServer := grpc.NewServer(serverOpts...)

	logger = logger.Session("grpc-server")

	db, err := parserOpts.SQL.Open()
	if err != nil {
		logger.Fatal(messages.ErrFailedToOpenSQLConnection, err)
		os.Exit(1)
	}

	pingLogger := logger.Session(messages.PingSQLConnection, parserOpts.SQL.LagerData())
	pingLogger.Debug(messages.Starting)
	err = db.Ping()
	if err != nil {
		logger.Fatal(messages.ErrFailedToPingSQLConnection, err, parserOpts.SQL.LagerData())
		os.Exit(1)
	}
	pingLogger.Debug(messages.Finished)

	defer db.Close()

	roleServiceServer := rpc.NewRoleServiceServer(logger, db)
	protos.RegisterRoleServiceServer(grpcServer, roleServiceServer)
	logger.Info(messages.StartingServer, listeningLogData)

	grpcServer.Serve(lis)
}

func (o *sqlOptions) Open() (*sql.DB, error) {
	switch o.DB.Driver {
	case "mysql":
		cfg := mysql.NewConfig()

		cfg.User = o.DB.Username
		cfg.Passwd = o.DB.Password
		cfg.Net = "tcp"
		cfg.Addr = net.JoinHostPort(o.DB.Host, strconv.Itoa(o.DB.Port))
		cfg.DBName = o.DB.Schema

		return sql.Open(o.DB.Driver, cfg.FormatDSN())
	default:
		return nil, errors.New("unsupported sql driver")
	}
}

func (o *sqlOptions) LagerData() lager.Data {
	return lager.Data{
		"db_driver":   o.DB.Driver,
		"db_host":     o.DB.Host,
		"db_port":     o.DB.Port,
		"db_schema":   o.DB.Schema,
		"db_username": o.DB.Username,
	}
}
