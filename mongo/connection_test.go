package mongo

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Connection", func() {
	Describe("Init", func() {
		When("URI and databaseName are correct", func() {
			It("sets the correct mongoDatabase global", func() {
				err := Init("mongodb://localhost:27017", "database")
				Expect(err).To(BeNil())
				Expect(mongoDatabase).NotTo(BeNil())
			})
		})

		When("URI is empty", func() {
			It("returns invalid URI errror", func() {
				err := Init("", "database")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ErrInvaidURI))
			})
		})

		When("URI is incorrect", func() {
			It("returns time out", func() {
				err := Init("mongodb://unknowhost:27017", "database")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ErrCouldNotConnect))
			})
		})

		When("DatabaseName is empty", func() {
			It("returns invalid DatabaseName error", func() {
				err := Init("mongodb://localhost:27017", "")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ErrInvaidDatabaseName))
			})
		})
	})

})
