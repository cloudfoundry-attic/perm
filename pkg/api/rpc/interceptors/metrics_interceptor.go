package interceptors

import (
	"context"
	"fmt"
	"time"

	"code.cloudfoundry.org/perm/pkg/metrics"
	"google.golang.org/grpc"
)

func MetricsInterceptor(statter metrics.Statter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		_, endpoint := parseFullMethod(info.FullMethod)

		statter.Inc(fmt.Sprintf("perm.count.%s", endpoint), 1)

		var successValue int64
		res, err := handler(ctx, req)

		if err == nil {
			successValue = 1
		}

		statter.Gauge(fmt.Sprintf("perm.success.%s", endpoint), successValue)

		end := time.Since(start)
		statter.TimingDuration(fmt.Sprintf("perm.requestduration.%s", endpoint), end)

		return res, err
	}
}
