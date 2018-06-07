package perm

import (
	"context"
	"crypto/tls"

	"code.cloudfoundry.org/perm/pkg/api/protos"
	"golang.org/x/oauth2"
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

type perRPCCredentials struct {
	tokenSource oauth2.TokenSource
}

func newPerRPCCredentials(tokenSource oauth2.TokenSource) *perRPCCredentials {
	return &perRPCCredentials{
		tokenSource: oauth2.ReuseTokenSource(nil, tokenSource),
	}
}

func (c *perRPCCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	token, err := c.tokenSource.Token()
	if err != nil {
		return nil, err
	}
	return map[string]string{"token": token.AccessToken}, nil
}

func (*perRPCCredentials) RequireTransportSecurity() bool {
	return true
}

func Dial(addr string, dialOpts ...DialOption) (*Client, error) {
	opts := &options{}

	for _, dialOpt := range dialOpts {
		dialOpt(opts)
	}

	var grpcOpts []grpc.DialOption

	if opts.transportCredentials != nil {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(opts.transportCredentials))
	} else {
		return nil, ErrNoTransportSecurity
	}

	if opts.tokenSource != nil {
		grpcOpts = append(grpcOpts, grpc.WithPerRPCCredentials(newPerRPCCredentials(opts.tokenSource)))
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
			Action:          p.Action,
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
		return Role{}, NewErrorFromStatus(s)
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
		return NewErrorFromStatus(s)
	}
}

func (c *Client) AssignRole(ctx context.Context, roleName string, actor Actor) error {
	req := &protos.AssignRoleRequest{
		RoleName: roleName,
		Actor: &protos.Actor{
			ID:        actor.ID,
			Namespace: actor.Namespace,
		},
	}
	_, err := c.roleServiceClient.AssignRole(ctx, req)
	s := status.Convert(err)
	switch s.Code() {
	case codes.OK:
		return nil
	default:
		return NewErrorFromStatus(s)
	}
}

func (c *Client) UnassignRole(ctx context.Context, roleName string, actor Actor) error {
	req := &protos.UnassignRoleRequest{
		RoleName: roleName,
		Actor: &protos.Actor{
			ID:        actor.ID,
			Namespace: actor.Namespace,
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
		return NewErrorFromStatus(s)
	}
}

func (c *Client) HasPermission(ctx context.Context, actor Actor, action, resource string) (bool, error) {
	req := &protos.HasPermissionRequest{
		Actor: &protos.Actor{
			ID:        actor.ID,
			Namespace: actor.Namespace,
		},
		Action:   action,
		Resource: resource,
	}
	res, err := c.permissionServiceClient.HasPermission(ctx, req)
	s := status.Convert(err)

	switch s.Code() {
	case codes.OK:
		return res.HasPermission, nil
	default:
		return false, NewErrorFromStatus(s)
	}
}

func (c *Client) ListResourcePatterns(ctx context.Context, actor Actor, action string) ([]string, error) {
	req := &protos.ListResourcePatternsRequest{
		Actor: &protos.Actor{
			ID:        actor.ID,
			Namespace: actor.Namespace,
		},
		Action: action,
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
		return nil, NewErrorFromStatus(s)
	}
}

type DialOption func(*options)

func WithTLSConfig(config *tls.Config) DialOption {
	return func(o *options) {
		o.transportCredentials = credentials.NewTLS(config)
	}
}

func WithTokenSource(tokenSource oauth2.TokenSource) DialOption {
	return func(o *options) {
		o.tokenSource = tokenSource
	}
}

type options struct {
	transportCredentials credentials.TransportCredentials
	tokenSource          oauth2.TokenSource
}
