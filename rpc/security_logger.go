package rpc

import "code.cloudfoundry.org/perm/logging"

//go:generate counterfeiter . SecurityLogger
type SecurityLogger interface {
	Log(signature logging.SecurityLoggerSignature, name logging.SecurityLoggerName)
}
