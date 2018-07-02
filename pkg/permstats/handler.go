package permstats

import (
	"context"
	"strconv"
	"strings"

	"code.cloudfoundry.org/perm/pkg/monitor"
	"google.golang.org/grpc/stats"
)

type Handler struct {
	statter monitor.PermStatter
}

const methodNameKey = "PermRPCMethodName"

func (h *Handler) TagRPC(c context.Context, info *stats.RPCTagInfo) context.Context {
	parts := strings.Split(info.FullMethodName, "/")
	methodName := parts[len(parts)-1]
	return context.WithValue(c, methodNameKey, methodName)
}

func (h *Handler) HandleRPC(c context.Context, rpcStats stats.RPCStats) {
	methodName, _ := c.Value(methodNameKey).(string)

	switch s := rpcStats.(type) {
	case *stats.InHeader:
		h.statter.Inc("count."+methodName, 1, 1)
	case *stats.End:
		h.statter.TimingDuration("rpcduration."+methodName, s.EndTime.Sub(s.BeginTime), 1)
		success := "0"
		if s.Error == nil {
			success = "1"
		}
		h.statter.Raw("success."+methodName, success, 1)
	case *stats.InPayload:
		h.statter.Raw("requestsize."+methodName, strconv.Itoa(s.Length), 1)
	case *stats.OutPayload:
		h.statter.Raw("responsesize."+methodName, strconv.Itoa(s.Length), 1)
	}
}

func (h *Handler) TagConn(c context.Context, info *stats.ConnTagInfo) context.Context {
	return c
}

func (h *Handler) HandleConn(context.Context, stats.ConnStats) {}

func NewHandler(statter monitor.PermStatter) *Handler {
	return &Handler{statter: statter}
}
