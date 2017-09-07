package rpc

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func togRPCError(err error) error {
	return status.Errorf(codes.Unknown, err.Error())
}
