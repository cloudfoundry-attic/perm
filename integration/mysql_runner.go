package integration

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"strconv"

	"time"

	"code.cloudfoundry.org/perm/cmd"
	_ "github.com/go-sql-driver/mysql"
)

// role and actor deletion cascade to assignments
var truncateStmts = []string{
	"DELETE FROM role",
	"DELETE FROM actor",
}

type ioReader struct{}

func (ioReader) ReadFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}

type statter struct{}

func (statter) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

type MySQLRunner struct {
	SQLFlag cmd.SQLFlag
}

func NewRunner(flag cmd.SQLFlag) *MySQLRunner {
	return &MySQLRunner{
		SQLFlag: flag,
	}
}

func (r *MySQLRunner) CreateTestDB() {
	createDB := exec.Command(
		"mysql",
		"--user", r.SQLFlag.DB.Username,
		fmt.Sprintf("--password=%s", r.SQLFlag.DB.Password),
		"--port", strconv.Itoa(r.SQLFlag.DB.Port),
		"-e", fmt.Sprintf("CREATE DATABASE %s", r.SQLFlag.DB.Schema),
	)
	session, err := gexec.Start(createDB, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	Eventually(session, 5*time.Second).Should(gexec.Exit(0))

	c := &cmd.UpCommand{
		Logger: cmd.LagerFlag{LogLevel: "error"},
		SQL:    r.SQLFlag,
	}

	err = c.Execute([]string{})
	Expect(err).NotTo(HaveOccurred())
}

func (r *MySQLRunner) DropTestDB() {
	dropDB := exec.Command(
		"mysql",
		"--user", r.SQLFlag.DB.Username,
		fmt.Sprintf("--password=%s", r.SQLFlag.DB.Password),
		"--port", strconv.Itoa(r.SQLFlag.DB.Port),
		"-e", fmt.Sprintf("DROP DATABASE %s", r.SQLFlag.DB.Schema),
	)
	session, err := gexec.Start(dropDB, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	Eventually(session, 5*time.Second).Should(gexec.Exit(0))
}

func (r *MySQLRunner) Truncate() {
	dbConn, err := r.SQLFlag.Open(statter{}, ioReader{})
	Expect(err).NotTo(HaveOccurred())

	for _, s := range truncateStmts {
		_, err = dbConn.Exec(s)
		Expect(err).NotTo(HaveOccurred())
	}
}
