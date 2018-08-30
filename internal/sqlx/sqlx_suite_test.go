package sqlx_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSqlx(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sqlx Suite")
}
