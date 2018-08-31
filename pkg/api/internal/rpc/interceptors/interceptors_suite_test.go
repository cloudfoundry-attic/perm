package interceptors_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
)

func TestInterceptors(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Interceptors Suite")
}

func testUnaryServerInterceptor(interceptorFactory func() grpc.UnaryServerInterceptor) {
	var (
		subject grpc.UnaryServerInterceptor
	)

	BeforeEach(func() {
		subject = interceptorFactory()
	})

	It("provides the expected arguments to the handler", func() {
		var (
			actualCtx context.Context
			actualReq interface{}
		)

		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			actualCtx = ctx
			actualReq = req

			return nil, nil
		}

		expectedReq := "test request"
		expectedCtxValue := "test context value"
		expectedCtx := context.WithValue(context.Background(), ctxKey{}, expectedCtxValue)

		_, err := subject(expectedCtx, expectedReq, &grpc.UnaryServerInfo{}, handler)
		Expect(err).NotTo(HaveOccurred())

		v := actualCtx.Value(ctxKey{})
		Expect(v).To(Equal(expectedCtxValue))

		Expect(actualReq).To(Equal(expectedReq))
	})

	It("returns the original result when the call succeeds", func() {
		expectedRes := "expected res"
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return expectedRes, nil
		}

		res, err := subject(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)

		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(Equal(expectedRes))
	})

	It("returns the original error code when the call fails", func() {
		testErr := errors.New("test error")
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, testErr
		}

		_, err := subject(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)

		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(testErr))
	})
}

type ctxKey struct{}
