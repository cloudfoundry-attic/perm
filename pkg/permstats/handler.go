package permstats

import (
	"context"
	"strings"

	"code.cloudfoundry.org/perm/pkg/metrics"
	"google.golang.org/grpc/stats"
)

const (
	methodNameKey = "PermRPCMethodName"
)

type Handler struct {
	statter metrics.Statter
}

func (h *Handler) TagRPC(c context.Context, info *stats.RPCTagInfo) context.Context {
	parts := strings.Split(info.FullMethodName, "/")
	methodName := parts[len(parts)-1]
	return context.WithValue(c, methodNameKey, methodName)
}

func (h *Handler) HandleRPC(c context.Context, rpcStats stats.RPCStats) {
	methodName, _ := c.Value(methodNameKey).(string)

	switch s := rpcStats.(type) {
	case *stats.InHeader:
		h.statter.Inc("perm.count."+methodName, 1)
	case *stats.End:
		h.statter.TimingDuration("perm.requestduration."+methodName, s.EndTime.Sub(s.BeginTime))
		success := int64(0)
		if s.Error == nil {
			success = 1
		}
		h.statter.Gauge("perm.success."+methodName, success)
	case *stats.InPayload:
		h.statter.Gauge("perm.requestsize."+methodName, int64(s.Length))
	case *stats.OutPayload:
		h.statter.Gauge("perm.responsesize."+methodName, int64(s.Length))
	}
}

// Not used, implemented to satisfy the stats.Handler interface
func (h *Handler) TagConn(c context.Context, info *stats.ConnTagInfo) context.Context {
	return c
}

// Not used, implemented to satisfy the stats.Handler interface
func (h *Handler) HandleConn(context.Context, stats.ConnStats) {}

func NewHandler(statter metrics.Statter) *Handler {
	return &Handler{statter: statter}
}
