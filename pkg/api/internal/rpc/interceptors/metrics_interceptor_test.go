package interceptors_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"

	. "code.cloudfoundry.org/perm/pkg/api/internal/rpc/interceptors"
	"code.cloudfoundry.org/perm/pkg/metrics/testmetrics"
)

var _ = Describe("MetricsInterceptor", func() {
	var (
		req      interface{}
		endpoint string
		info     *grpc.UnaryServerInfo

		statter *testmetrics.Statter

		subject grpc.UnaryServerInterceptor
	)

	BeforeEach(func() {
		req = uuid.NewV4().String()
		service := fmt.Sprintf("%s.%s", uuid.NewV4().String(), uuid.NewV4().String())
		endpoint = uuid.NewV4().String()
		info = &grpc.UnaryServerInfo{
			FullMethod: fmt.Sprintf("/%s/%s", service, endpoint),
		}

		statter = testmetrics.NewStatter()

		subject = MetricsInterceptor(statter)
	})

	testUnaryServerInterceptor(func() grpc.UnaryServerInterceptor { return subject })

	It("increments the call count", func() {
		handler := func(context.Context, interface{}) (interface{}, error) {
			return nil, nil
		}

		_, _ = subject(context.Background(), req, info, handler)
		Expect(statter.IncCalls()).To(ContainElement(testmetrics.IncCall{
			Metric: fmt.Sprintf("perm.count.%s", endpoint),
			Value:  1,
		}))
	})

	It("records the call duration", func() {
		start := time.Now()
		handler := func(context.Context, interface{}) (interface{}, error) {
			return nil, nil
		}

		_, _ = subject(context.Background(), req, info, handler)

		end := time.Since(start)

		Expect(statter.TimingDurationCalls()).To(ContainElement(gstruct.MatchAllFields(gstruct.Fields{
			"Metric": Equal(fmt.Sprintf("perm.requestduration.%s", endpoint)),
			"Value":  BeNumerically(">", 0),
		})))
		Expect(statter.TimingDurationCalls()).To(ContainElement(gstruct.MatchAllFields(gstruct.Fields{
			"Metric": Equal(fmt.Sprintf("perm.requestduration.%s", endpoint)),
			"Value":  BeNumerically("<=", end),
		})))
	})

	It("records success when the request succeeds", func() {
		handler := func(context.Context, interface{}) (interface{}, error) {
			return nil, nil
		}

		_, _ = subject(context.Background(), req, info, handler)
		Expect(statter.GaugeCalls()).To(ContainElement(testmetrics.GaugeCall{
			Metric: fmt.Sprintf("perm.success.%s", endpoint),
			Value:  1,
		}))
	})

	It("records failure when the request fails", func() {
		handler := func(context.Context, interface{}) (interface{}, error) {
			return nil, errors.New("test error")
		}

		_, _ = subject(context.Background(), req, info, handler)
		Expect(statter.GaugeCalls()).To(ContainElement(testmetrics.GaugeCall{
			Metric: fmt.Sprintf("perm.success.%s", endpoint),
			Value:  0,
		}))
	})
})
