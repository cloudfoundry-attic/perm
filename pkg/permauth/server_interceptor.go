package permauth

import (
	"context"

	"code.cloudfoundry.org/perm"
	"code.cloudfoundry.org/perm/logx"
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

const (
	AuthFailSignature = "AuthFail"
	AuthPassSignature = "AuthPass"
)

type ctxKey struct{}

type User struct {
	ID string
}

func NewUserContext(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, ctxKey{}, user)
}

func UserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(ctxKey{}).(User)
	return user, ok
}

func ServerInterceptor(provider OIDCProvider, securityLogger logx.SecurityLogger) grpc.UnaryServerInterceptor {
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

		user := User{
			ID: idToken.Subject,
		}

		return handler(NewUserContext(ctx, user), req)
	}
}
