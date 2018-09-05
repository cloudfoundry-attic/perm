package monitor_test

import (
	"time"

	. "code.cloudfoundry.org/perm/monitor"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("Probe", func() {
	var (
		defaultTimeout        = time.Second
		defaultCleanupTimeout = time.Second * 10
		defaultMaxLatency     = time.Millisecond * 100
	)

	Describe("with no options provided", func() {
		testProbe(defaultTimeout, defaultCleanupTimeout, defaultMaxLatency)
	})

	Describe("with a timeout provided", func() {
		timeout := defaultTimeout * 2
		testProbe(timeout, defaultCleanupTimeout, defaultMaxLatency, WithTimeout(timeout))
	})

	Describe("with a cleanup timeout provided", func() {
		cleanupTimeout := defaultCleanupTimeout * 2
		testProbe(defaultTimeout, cleanupTimeout, defaultMaxLatency, WithCleanupTimeout(cleanupTimeout))
	})

	Describe("with a max latency provided", func() {
		maxLatency := defaultMaxLatency * 2
		testProbe(defaultTimeout, defaultCleanupTimeout, maxLatency, WithMaxLatency(maxLatency))
	})
})
