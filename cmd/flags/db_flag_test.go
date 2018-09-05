package flags_test

import (
	"context"

	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/perm/cmd/flags"
	"code.cloudfoundry.org/perm/logx"
	"code.cloudfoundry.org/perm/logx/lagerx"
)

var _ = Describe("DBFlag", func() {
	var (
		ctx    context.Context
		logger logx.Logger

		flag *flags.DBFlag
	)

	BeforeEach(func() {
		ctx = context.Background()
		logger = lagerx.NewLogger(lagertest.NewTestLogger("flags"))

		flag = &flags.DBFlag{
			Driver:   "mysql",
			Host:     "localhost",
			Port:     1234,
			Schema:   "perm",
			Username: "perm-user",
			Password: "perm-password",
		}
	})

	Describe("an in-memory connection", func() {
		It("does not require all DB arguments", func() {
			memFlag := &flags.DBFlag{
				Driver: "in-memory",
			}

			_, err := memFlag.Connect(ctx, logger)
			Expect(err).To(MatchError("Connect() unsupported for in-memory driver"))
		})
	})

	Describe("a connection to a real database", func() {
		It("requires a host", func() {
			flag.Host = ""

			_, err := flag.Connect(ctx, logger)
			Expect(err).To(MatchError("the required host parameter was not specified; see --help"))
		})

		It("requires a port", func() {
			flag.Port = 0

			_, err := flag.Connect(ctx, logger)
			Expect(err).To(MatchError("the required port parameter was not specified; see --help"))
		})

		It("requires a schema", func() {
			flag.Schema = ""

			_, err := flag.Connect(ctx, logger)
			Expect(err).To(MatchError("the required schema parameter was not specified; see --help"))
		})

		It("requires a username", func() {
			flag.Username = ""

			_, err := flag.Connect(ctx, logger)
			Expect(err).To(MatchError("the required username parameter was not specified; see --help"))
		})
	})
})
