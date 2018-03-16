package rpc

import (
	"code.cloudfoundry.org/perm/logging"
	"context"
)

//go:generate counterfeiter . SecurityLogger
type SecurityLogger interface {
	Log(ctx context.Context, signature, name string, args ...logging.CustomExtension)
}
