package rpc

import (
	"context"

	"code.cloudfoundry.org/perm/pkg/api/logging"
)

//go:generate counterfeiter . SecurityLogger
type SecurityLogger interface {
	Log(ctx context.Context, signature, name string, args ...logging.CustomExtension)
}
