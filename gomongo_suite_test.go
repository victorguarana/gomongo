package gomongo_test

import (
	"testing"

	"github.com/onsi/gomega/format"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGomongo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gomongo Suite")
}

var _ = BeforeSuite(func() {
	limitGomegaMaxLenght()
	removeTestContainerLogs()
})

func limitGomegaMaxLenght() {
	format.MaxLength = 300
}
