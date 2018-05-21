package cmd_test

import (
	. "code.cloudfoundry.org/perm/cmd"
	"code.cloudfoundry.org/perm/cmd/flags"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("perm migrate", func() {
	Describe("DownCommand", func() {
		var downCmd DownCommand
		It("performs a no-op down-migration for in-memory driver", func() {
			downCmd = DownCommand{
				Logger: flags.LagerFlag{LogLevel: "fatal"},
				DB: flags.DBFlag{
					Driver: "in-memory",
				},
			}
			err := downCmd.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())
		})
		It("performs errors out on unsupported driver", func() {
			downCmd = DownCommand{
				Logger: flags.LagerFlag{LogLevel: "fatal"},
				DB: flags.DBFlag{
					Driver:   "unsupported-driver",
					Host:     "host",
					Port:     2313,
					Schema:   "perm",
					Username: "perm",
					Password: "perm",
				},
			}
			err := downCmd.Execute([]string{})
			Expect(err).To(MatchError("unsupported sql driver"))
		})
	})
	Describe("UpCommand", func() {
		var upCmd UpCommand
		It("performs a no-op up-migration for in-memory driver", func() {
			upCmd = UpCommand{
				Logger: flags.LagerFlag{LogLevel: "fatal"},
				DB: flags.DBFlag{
					Driver: "in-memory",
				},
			}
			err := upCmd.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())
		})
		It("performs errors out on unsupported driver", func() {
			upCmd = UpCommand{
				Logger: flags.LagerFlag{LogLevel: "fatal"},
				DB: flags.DBFlag{
					Driver:   "unsupported-driver",
					Host:     "host",
					Port:     2313,
					Schema:   "perm",
					Username: "perm",
					Password: "perm",
				},
			}
			err := upCmd.Execute([]string{})
			Expect(err).To(MatchError("unsupported sql driver"))
		})
	})

})
