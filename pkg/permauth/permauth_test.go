package permauth_test

import (
	"context"

	"code.cloudfoundry.org/perm/pkg/api/rpc/rpcfakes"
	"code.cloudfoundry.org/perm/pkg/permauth"
	"code.cloudfoundry.org/perm/pkg/permauth/permauthfakes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Auth Server Interceptor", func() {
	var (
		interceptor        grpc.UnaryServerInterceptor
		fakeProvider       *permauthfakes.FakeOIDCProvider
		fakeSecurityLogger *rpcfakes.FakeSecurityLogger
		sampleHandler      func(context.Context, interface{}) (interface{}, error)
	)

	BeforeEach(func() {
		fakeProvider = new(permauthfakes.FakeOIDCProvider)
		fakeSecurityLogger = new(rpcfakes.FakeSecurityLogger)
		interceptor = permauth.ServerInterceptor(fakeProvider, fakeSecurityLogger)
		sampleHandler = func(c context.Context, r interface{}) (interface{}, error) { return nil, nil }
	})

	It("errors out when context contains no metadata", func() {
		interceptor(context.Background(), nil, nil, sampleHandler)

		Expect(fakeSecurityLogger.LogCallCount()).To(Equal(1))
		_, logID, logName, extension := fakeSecurityLogger.LogArgsForCall(0)
		Expect(logID).To(Equal("Auth"))
		Expect(logName).To(Equal("Auth"))
		Expect(extension).To(HaveLen(1))
		Expect(extension[0].Value).To(ContainSubstring("unexpected: cannot extract metadata from context"))
	})

	It("errors out when context does not contain a token field", func() {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{"key": "value"}))
		interceptor(ctx, nil, nil, sampleHandler)

		Expect(fakeSecurityLogger.LogCallCount()).To(Equal(1))
		_, logID, logName, extension := fakeSecurityLogger.LogArgsForCall(0)
		Expect(logID).To(Equal("Auth"))
		Expect(logName).To(Equal("Auth"))
		Expect(extension).To(HaveLen(1))
		Expect(extension[0].Value).To(ContainSubstring("token field not in the metadata"))
	})

	It("errors out when token isn't valid", func() {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{"token": "hello"}))
		interceptor(ctx, nil, nil, sampleHandler)

		Expect(fakeSecurityLogger.LogCallCount()).To(Equal(1))
		_, logID, logName, extension := fakeSecurityLogger.LogArgsForCall(0)
		Expect(logID).To(Equal("Auth"))
		Expect(logName).To(Equal("Auth"))
		Expect(extension).To(HaveLen(1))
		Expect(extension[0].Value).To(ContainSubstring("oidc: malformed jwt"))
	})
})
