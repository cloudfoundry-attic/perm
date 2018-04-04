package perm

import (
	"crypto/tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Client struct {
	conn *grpc.ClientConn
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

	return &Client{
		conn: conn,
	}, nil
}

func (c *Client) Close() error {
	if err := c.conn.Close(); err != nil {
		return ErrClientConnClosing
	}
	return nil
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
