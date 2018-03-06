package rpc

import "context"

//go:generate counterfeiter . SecurityLogger
type SecurityLogger interface {
	Log(ctx context.Context, signature, name string)
}
