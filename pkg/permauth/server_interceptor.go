package permauth

import (
	"context"

	"code.cloudfoundry.org/perm/pkg/logx"
	"code.cloudfoundry.org/perm/pkg/logx/cef"
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

func ServerInterceptor(provider OIDCProvider, securityLogger logx.SecurityLogger) grpc.UnaryServerInterceptor {
	verifier := provider.Verifier(&oidc.Config{
		ClientID: "perm",
	})

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			securityLogger.Log(ctx, "Auth", "Auth", cef.CustomExtension{Key: "msg", Value: "no metadata"})
			return nil, perm.ErrUnauthenticated
		}

		token, ok := md["token"]
		if !ok {
			securityLogger.Log(ctx, "Auth", "Auth", cef.CustomExtension{Key: "msg", Value: "no token"})
			return nil, perm.ErrUnauthenticated
		}

		idToken, err := verifier.Verify(ctx, token[0])
		if err != nil {
			securityLogger.Log(ctx, "Auth", "Auth", cef.CustomExtension{Key: "msg", Value: err.Error()})
			return nil, perm.ErrUnauthenticated
		}

		extensions := []cef.CustomExtension{
			cef.CustomExtension{Key: "msg", Value: "authentication succeeded"},
			cef.CustomExtension{Key: "subject", Value: idToken.Subject},
		}
		securityLogger.Log(ctx, "Auth", "Auth", extensions...)
		return handler(ctx, req)
	}
}
