package perm

import (
	"context"
	"crypto/tls"

	"code.cloudfoundry.org/perm-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

type Client struct {
	conn *grpc.ClientConn

	roleServiceClient       protos.RoleServiceClient
	permissionServiceClient protos.PermissionServiceClient
}

func Dial(addr string, opts ...DialOption) (*Client, error) {
	config := &options{}

	for _, opt := range opts {
		opt(config)
	}

	var grpcOpts []grpc.DialOption

	if config.transportCredentials != nil {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(config.transportCredentials))
	} else {
		return nil, ErrNoTransportSecurity
	}

	conn, err := grpc.Dial(addr, grpcOpts...)
	if err != nil {
		return nil, ErrFailedToConnect
	}

	roleServiceClient := protos.NewRoleServiceClient(conn)
	permissionServiceClient := protos.NewPermissionServiceClient(conn)

	return &Client{
		conn:                    conn,
		roleServiceClient:       roleServiceClient,
		permissionServiceClient: permissionServiceClient,
	}, nil
}

func (c *Client) Close() error {
	if err := c.conn.Close(); err != nil {
		return ErrClientConnClosing
	}
	return nil
}

func (c *Client) CreateRole(ctx context.Context, name string, permissions ...Permission) (Role, error) {
	var reqPermissions []*protos.Permission

	for _, p := range permissions {
		reqPermissions = append(reqPermissions, &protos.Permission{
			Name:            p.Action,
			ResourcePattern: p.ResourcePattern,
		})
	}

	req := &protos.CreateRoleRequest{
		Name:        name,
		Permissions: reqPermissions,
	}

	res, err := c.roleServiceClient.CreateRole(ctx, req)
	s := status.Convert(err)

	switch s.Code() {
	case codes.OK:
		return Role{
			Name: res.GetRole().GetName(),
		}, nil
	case codes.AlreadyExists:
		return Role{}, ErrRoleAlreadyExists
	default:
		return Role{}, ErrUnknown
	}
}

func (c *Client) DeleteRole(ctx context.Context, name string) error {
	req := &protos.DeleteRoleRequest{
		Name: name,
	}
	_, err := c.roleServiceClient.DeleteRole(ctx, req)
	s := status.Convert(err)

	switch s.Code() {
	case codes.OK:
		return nil
	case codes.NotFound:
		return ErrRoleNotFound
	default:
		return ErrUnknown
	}
}

type DialOption func(*options)

func WithTLSConfig(config *tls.Config) DialOption {
	return func(o *options) {
		o.transportCredentials = credentials.NewTLS(config)
	}
}

type options struct {
	transportCredentials credentials.TransportCredentials
}
