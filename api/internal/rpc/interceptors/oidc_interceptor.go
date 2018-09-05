package interceptors

import (
	"context"

	"code.cloudfoundry.org/perm"
	"code.cloudfoundry.org/perm/internal/models"
	"code.cloudfoundry.org/perm/logx"
	"code.cloudfoundry.org/perm/oidcx"
	oidc "github.com/coreos/go-oidc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const permAdminScope = "perm.admin"

type Claims struct {
	Scopes []string `json:"scope"`
}

const (
	AuthFailSignature = "AuthFail"
	AuthPassSignature = "AuthPass"
)

func OIDCInterceptor(provider oidcx.Provider, securityLogger logx.SecurityLogger) grpc.UnaryServerInterceptor {
	verifier := provider.Verifier(&oidc.Config{
		ClientID: "perm",
	})

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			securityLogger.Log(ctx, AuthFailSignature, "missing token", logx.SecurityData{Key: "msg", Value: "no metadata"})
			return nil, perm.ErrUnauthenticated
		}

		token, ok := md["token"]
		if !ok {
			securityLogger.Log(ctx, AuthFailSignature, "missing token", logx.SecurityData{Key: "msg", Value: "no token"})
			return nil, perm.ErrUnauthenticated
		}

		idToken, err := verifier.Verify(ctx, token[0])
		if err != nil {
			securityLogger.Log(ctx, AuthFailSignature, "invalid token", logx.SecurityData{Key: "msg", Value: err.Error()})
			return nil, perm.ErrUnauthenticated
		}

		extensions := []logx.SecurityData{
			logx.SecurityData{Key: "msg", Value: "auth succeeded"},
			logx.SecurityData{Key: "subject", Value: idToken.Subject},
		}
		securityLogger.Log(ctx, AuthPassSignature, "auth succeeded", extensions...)

		user := models.User{
			ID: idToken.Subject,
		}

		return handler(models.NewUserContext(ctx, user), req)
	}
}
