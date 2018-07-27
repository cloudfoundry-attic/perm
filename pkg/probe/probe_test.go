package probe_test

import (
	"time"

	. "code.cloudfoundry.org/perm/pkg/probe"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("Probe", func() {
	var (
		defaultTimeout        = time.Second
		defaultCleanupTimeout = time.Second * 10
		defaultMaxLatency     = time.Millisecond * 100
	)

	Describe("with no timeout provided", func() {
		testProbe(defaultTimeout, defaultCleanupTimeout, defaultMaxLatency)
	})

	Describe("with a timeout provided", func() {
		timeout := time.Hour

		testProbe(timeout, defaultCleanupTimeout, defaultMaxLatency, WithTimeout(timeout))
	})

	Describe("with a cleanup timeout provided", func() {
		cleanupTimeout := time.Hour

		testProbe(defaultTimeout, cleanupTimeout, defaultMaxLatency, WithCleanupTimeout(cleanupTimeout))
	})

	Describe("with a max latency provided", func() {
		maxLatency := time.Minute

		testProbe(defaultTimeout, defaultCleanupTimeout, maxLatency, WithMaxLatency(maxLatency))
	})
})
