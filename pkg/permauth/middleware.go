package permauth

import (
	"context"

	"code.cloudfoundry.org/perm/pkg/perm"
	"google.golang.org/grpc"
)

func Middleware(requireAuth bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if requireAuth {
			return nil, perm.ErrUnauthenticated
		}
		return handler(ctx, req)
	}
}
