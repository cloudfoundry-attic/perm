package rpc

import (
	"code.cloudfoundry.org/perm/errdefs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func togRPCError(code codes.Code, err error) error {
	return status.Errorf(code, err.Error())
}

func togRPCErrorNew(err error) error {
	switch err.(type) {
	case errdefs.ErrNotFound:
		return status.Errorf(codes.NotFound, err.Error())
	case errdefs.ErrAlreadyExists:
		return status.Errorf(codes.AlreadyExists, err.Error())
	default:
		return status.Errorf(codes.Unknown, err.Error())
	}
}
