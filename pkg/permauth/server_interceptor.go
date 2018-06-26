package permauth

import (
	"context"

	"code.cloudfoundry.org/perm/pkg/api/rpc"
	"code.cloudfoundry.org/perm/pkg/perm"
	oidc "github.com/coreos/go-oidc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const permAdminScope = "perm.admin"

//go:generate counterfeiter . OIDCProvider

type OIDCProvider interface {
	Verifier(config *oidc.Config) *oidc.IDTokenVerifier
}

type Claims struct {
	Scopes []string `json:"scope"`
}

func ServerInterceptor(provider OIDCProvider, securityLogger rpc.SecurityLogger) grpc.UnaryServerInterceptor {
	verifier := provider.Verifier(&oidc.Config{
		ClientID: "perm",
	})

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, perm.ErrUnauthenticated
		}

		token, ok := md["token"]
		if !ok {
			return nil, perm.ErrUnauthenticated
		}

		_, err = verifier.Verify(ctx, token[0])
		if err != nil {
			return nil, perm.ErrUnauthenticated
		}

		return handler(ctx, req)
	}
}
