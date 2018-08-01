package monitor

import (
	"time"
)

const (
	DefaultTimeout        = time.Second
	DefaultCleanupTimeout = time.Second * 10
	DefaultMaxLatency     = time.Millisecond * 100
)

type options struct {
	timeout        time.Duration
	cleanupTimeout time.Duration
	maxLatency     time.Duration
}

func defaultOptions() *options {
	return &options{
		timeout:        DefaultTimeout,
		cleanupTimeout: DefaultCleanupTimeout,
		maxLatency:     DefaultMaxLatency,
	}
}

type Option func(*options)

func WithTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.timeout = timeout
	}
}

func WithCleanupTimeout(cleanupTimeout time.Duration) Option {
	return func(o *options) {
		o.cleanupTimeout = cleanupTimeout
	}
}

func WithMaxLatency(latency time.Duration) Option {
	return func(o *options) {
		o.maxLatency = latency
	}
}
