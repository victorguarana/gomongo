package gomongo_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGomongo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gomongo Suite")
}
