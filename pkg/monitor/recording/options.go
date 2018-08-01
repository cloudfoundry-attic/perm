package recording

import "code.cloudfoundry.org/clock"

type options struct {
	clock clock.Clock
}

func defaultOptions() *options {
	return &options{
		clock: clock.NewClock(),
	}
}

type Option func(*options)

func WithClock(c clock.Clock) Option {
	return func(o *options) {
		o.clock = c
	}
}
