package cmd_test

import (
	. "code.cloudfoundry.org/perm/cmd"

	"errors"

	"code.cloudfoundry.org/perm/cmd/cmdfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FileOrStringFlag", func() {
	Describe("#Bytes", func() {
		var (
			statter *cmdfakes.FakeStatter
			reader  *cmdfakes.FakeFileReader
			info    *cmdfakes.FakeFileInfo

			subject FileOrStringFlag
		)

		BeforeEach(func() {
			statter = new(cmdfakes.FakeStatter)
			reader = new(cmdfakes.FakeFileReader)
			info = new(cmdfakes.FakeFileInfo)
		})

		It("returns the file contents if readable", func() {
			subject = "/some/fake/file"

			info.IsDirReturns(false)
			reader.ReadFileReturns([]byte("file contents"), nil)
			statter.StatReturns(info, nil)

			b, err := subject.Bytes(statter, reader)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(b)).To(Equal("file contents"))
		})

		It("returns the string if provided a string", func() {
			subject = "some string"

			statter.StatReturns(nil, errors.New("does not exist"))

			b, err := subject.Bytes(statter, reader)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(b)).To(Equal("some string"))
		})

		It("decodes the newlines if passed a string", func() {
			subject = "some\\nstring"
			expected := `some
string`

			statter.StatReturns(nil, errors.New("does not exist"))

			b, err := subject.Bytes(statter, reader)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(b)).To(Equal(expected))
		})

		It("fails if the path points to a directory", func() {
			subject = "/some/fake/dir"

			info.IsDirReturns(true)
			statter.StatReturns(info, nil)

			_, err := subject.Bytes(statter, reader)
			Expect(err).To(HaveOccurred())
		})

		It("fails if the file is not readable", func() {
			subject = "/some/fake/dir"

			info.IsDirReturns(false)
			reader.ReadFileReturns(nil, errors.New("error reading file"))
			statter.StatReturns(info, nil)

			_, err := subject.Bytes(statter, reader)
			Expect(err).To(MatchError("error reading file"))
		})
	})
})
