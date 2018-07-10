package logx

import (
	"context"
)

//go:generate counterfeiter . SecurityLogger

type SecurityData struct {
	Key   string
	Value string
}

type SecurityLogger interface {
	Log(ctx context.Context, signature, name string, args ...SecurityData)
}
