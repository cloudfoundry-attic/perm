package permstats

import (
	"context"
	"strings"

	"github.com/cactus/go-statsd-client/statsd"
	"google.golang.org/grpc/stats"
)

const (
	methodNameKey   = "PermRPCMethodName"
	statsSampleRate = 1
)

type Handler struct {
	statter statsd.Statter
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
		h.statter.Inc("perm.count."+methodName, 1, statsSampleRate)
	case *stats.End:
		h.statter.TimingDuration("perm.requesttime."+methodName, s.EndTime.Sub(s.BeginTime), statsSampleRate)
		success := int64(0)
		if s.Error == nil {
			success = 1
		}
		h.statter.Inc("perm.success."+methodName, success, statsSampleRate)
	case *stats.InPayload:
		h.statter.Inc("perm.requestsize."+methodName, int64(s.Length), statsSampleRate)
	case *stats.OutPayload:
		h.statter.Inc("perm.responsesize."+methodName, int64(s.Length), statsSampleRate)
	}
}

// Not used, implemented to satisfy the stats.Handler interface
func (h *Handler) TagConn(c context.Context, info *stats.ConnTagInfo) context.Context {
	return c
}

// Not used, implemented to satisfy the stats.Handler interface
func (h *Handler) HandleConn(context.Context, stats.ConnStats) {}

func NewHandler(statter statsd.Statter) *Handler {
	return &Handler{statter: statter}
}
