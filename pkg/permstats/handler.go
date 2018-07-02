package permstats

import (
	"context"

	"google.golang.org/grpc/stats"
)

type Handler struct {
}

func (h *Handler) TagRPC(c context.Context, info *stats.RPCTagInfo) context.Context {
	return c
}
func (h *Handler) HandleRPC(c context.Context, stats stats.RPCStats) {
}
func (h *Handler) TagConn(c context.Context, info *stats.ConnTagInfo) context.Context {
	return c
}
func (h *Handler) HandleConn(context.Context, stats.ConnStats) {}
