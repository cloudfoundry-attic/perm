package ioutilx

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"io/ioutil"
	"os"
)

var _ = Describe("ioutil", func() {
	Describe("#OpenLogFile", func() {
		var dirName, logFilePath string
		var err error
		BeforeEach(func() {
			dirName, err = ioutil.TempDir("", "perm-test")
			Expect(err).NotTo(HaveOccurred())
			logFilePath = dirName + "/audit.log"
		})
		AfterEach(func() {
			//Expect(os.RemoveAll(dirName)).NotTo(HaveOccurred())
		})
		It("creates a non-existent audit file", func() {
			file, err := OpenLogFile(logFilePath)
			Expect(err).NotTo(HaveOccurred())

			defer file.Close()

			fileInfo, err := os.Stat(logFilePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(fileInfo.Mode()).To(Equal(os.FileMode(0600)))
			Expect(fileInfo.Name()).To(Equal("audit.log"))
		})

		It("appends to an existing audit file", func() {
			err := ioutil.WriteFile(logFilePath, []byte("logline1\nlogline2\n"), 0600)
			Expect(err).NotTo(HaveOccurred())
			logFile, err := OpenLogFile(logFilePath)
			_, err = logFile.Write([]byte("logline3\n"))
			Expect(err).NotTo(HaveOccurred())
			logFile.Close()

			contents, err := ioutil.ReadFile(logFilePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(ContainSubstring("logline1\nlogline2\nlogline3\n"))
		})

		Context("when the directory does not exist", func() {
			BeforeEach(func() {
				Expect(os.RemoveAll(dirName)).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				_, err := OpenLogFile("/nonexistent/audit.log")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no such file or directory"))
			})
		})
	})
})
