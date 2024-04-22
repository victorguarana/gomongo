package gomongo

import (
	"context"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/mongodb"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NewDatabase", Ordered, func() {
	var (
		mongodbContainer    *mongodb.MongoDBContainer
		mongodbContainerURI string
		connectionSettings  ConnectionSettings
	)

	BeforeAll(func() {
		mongodbContainer, mongodbContainerURI = runMongoContainer(context.Background())
		connectionSettings = ConnectionSettings{
			URI:               mongodbContainerURI,
			DatabaseName:      "test",
			ConnectionTimeout: 1 * time.Second,
		}
	})

	Context("when mongo is up", func() {
		Context("when mongo is running", func() {
			It("returns nil", func() {
				receivedDatabase, receivedErr := NewDatabase(connectionSettings)
				Expect(receivedErr).NotTo(HaveOccurred())
				Expect(receivedDatabase.mongoDatabase).NotTo(BeNil())
			})
		})

	})

	Context("when mongo is down", func() {
		BeforeAll(func() {
			terminateMongoContainer(mongodbContainer, context.Background())
		})

		It("returns error", func() {
			receivedDatabase, receivedErr := NewDatabase(connectionSettings)
			Expect(receivedErr).To(MatchError(ErrGomongoCanNotConnect))
			Expect(receivedDatabase).To(BeNil())
		})
	})
})
