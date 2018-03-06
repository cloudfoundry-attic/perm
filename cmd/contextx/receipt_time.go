package contextx

import (
	"context"
	"time"
)

type receiptTimeKey struct{}

func WithReceiptTime(parent context.Context, rt time.Time) context.Context {
	return context.WithValue(parent, receiptTimeKey{}, rt)
}

func ReceiptTimeFromContext(ctx context.Context) (time.Time, bool) {
	t, ok := ctx.Value(receiptTimeKey{}).(time.Time)
	return t, ok
}
