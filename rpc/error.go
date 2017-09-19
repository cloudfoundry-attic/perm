package rpc

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func togRPCError(code codes.Code, err error) error {
	return status.Errorf(code, err.Error())
}
