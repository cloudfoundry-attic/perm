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
	} else if config.insecure {
		grpcOpts = append(grpcOpts, grpc.WithInsecure())
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

func (c *Client) AssignRole(ctx context.Context, roleName string, actor Actor) error {
	req := &protos.AssignRoleRequest{
		RoleName: roleName,
		Actor: &protos.Actor{
			ID:     actor.ID,
			Issuer: actor.Namespace,
		},
	}
	_, err := c.roleServiceClient.AssignRole(ctx, req)
	s := status.Convert(err)

	switch s.Code() {
	case codes.OK:
		return nil
	case codes.NotFound:
		return ErrRoleNotFound
	case codes.AlreadyExists:
		return ErrAssignmentAlreadyExists
	default:
		return ErrUnknown
	}
}

func (c *Client) UnassignRole(ctx context.Context, roleName string, actor Actor) error {
	req := &protos.UnassignRoleRequest{
		RoleName: roleName,
		Actor: &protos.Actor{
			ID:     actor.ID,
			Issuer: actor.Namespace,
		},
	}
	_, err := c.roleServiceClient.UnassignRole(ctx, req)
	s := status.Convert(err)

	switch s.Code() {
	case codes.OK:
		return nil
	case codes.NotFound:
		return ErrAssignmentNotFound
	default:
		return ErrUnknown
	}
}

func (c *Client) HasPermission(ctx context.Context, actor Actor, action, resourceID string) (bool, error) {
	req := &protos.HasPermissionRequest{
		Actor: &protos.Actor{
			ID:     actor.ID,
			Issuer: actor.Namespace,
		},
		PermissionName: action,
		ResourceId:     resourceID,
	}
	res, err := c.permissionServiceClient.HasPermission(ctx, req)
	s := status.Convert(err)

	switch s.Code() {
	case codes.OK:
		return res.HasPermission, nil
	default:
		return false, ErrUnknown
	}
}

func (c *Client) ListResourcePatterns(ctx context.Context, actor Actor, action string) ([]string, error) {
	req := &protos.ListResourcePatternsRequest{
		Actor: &protos.Actor{
			ID:     actor.ID,
			Issuer: actor.Namespace,
		},
		PermissionName: action,
	}
	res, err := c.permissionServiceClient.ListResourcePatterns(ctx, req)
	s := status.Convert(err)

	switch s.Code() {
	case codes.OK:
		var resourcePatterns []string
		for _, resourcePattern := range res.GetResourcePatterns() {
			resourcePatterns = append(resourcePatterns, resourcePattern)
		}

		return resourcePatterns, nil
	default:
		return nil, ErrUnknown
	}
}

type DialOption func(*options)

func WithTLSConfig(config *tls.Config) DialOption {
	return func(o *options) {
		o.transportCredentials = credentials.NewTLS(config)
	}
}

func WithInsecure() DialOption {
	return func(o *options) {
		o.insecure = true
	}
}

type options struct {
	transportCredentials credentials.TransportCredentials
	insecure             bool
}
