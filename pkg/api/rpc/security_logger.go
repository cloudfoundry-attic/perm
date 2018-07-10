package rpc

import (
	"context"

	"code.cloudfoundry.org/perm/pkg/logx/cef"
)

//go:generate counterfeiter . SecurityLogger
type SecurityLogger interface {
	Log(ctx context.Context, signature, name string, args ...cef.CustomExtension)
}
