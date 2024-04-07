package gomongo

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConnectionSettings{}", func() {
	Describe("validate", func() {
		Describe("valid cases", func() {
			Context("when all settings are filled", func() {
				It("returns nil", func() {
					cs := ConnectionSettings{
						URI:               "mongodb://localhost:27017/test",
						DatabaseName:      "test",
						ConnectionTimeout: 10 * time.Second,
					}

					Expect(cs.validate()).To(Succeed())
				})
			})

			Context("when timeout is empty", func() {
				It("returns nil", func() {
					cs := ConnectionSettings{
						URI:          "mongodb://localhost:27017/test",
						DatabaseName: "test",
					}

					Expect(cs.validate()).To(Succeed())
				})
			})
		})

		Describe("invalid cases", func() {
			Context("when URI is empty", func() {
				It("returns ErrInvalidSettings", func() {
					cs := ConnectionSettings{
						DatabaseName: "test",
					}

					receivedErr := cs.validate()

					Expect(receivedErr).To(MatchError(ErrInvalidSettings))
					Expect(receivedErr).To(MatchError(ContainSubstring("URI can not be empty")))
				})
			})

			Context("when DatabaseName is empty", func() {
				It("returns ErrInvalidSettings", func() {
					cs := ConnectionSettings{
						URI: "mongodb://localhost:27017/test",
					}

					receivedErr := cs.validate()

					Expect(receivedErr).To(MatchError(ErrInvalidSettings))
					Expect(receivedErr).To(MatchError(ContainSubstring("Database Name can not be empty")))
				})
			})
		})

	})
})
