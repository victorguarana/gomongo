package gomongo

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
)

type DummyDocument struct {
	ID string `bson:"_id"`
}

var _ = Describe("NewCollection", Ordered, func() {
	var (
		gomongoDatabaseName   = "database_test"
		gomongoCollectionName = "collection_test"

		mongodbContainerURI string
		mongodbContainer    *mongodb.MongoDBContainer

		gomongoDatabase *Database
	)

	BeforeAll(func() {
		mongodbContainer, mongodbContainerURI = runMongoContainer(context.Background())
		gomongoDatabase, _ = NewDatabase(ConnectionSettings{
			URI:               mongodbContainerURI,
			DatabaseName:      gomongoDatabaseName,
			ConnectionTimeout: time.Second,
		})
	})

	Context("when database is initialized", func() {
		It("returns collection", func() {
			receivedCollection, receivedErr := NewCollection[DummyDocument](gomongoDatabase, gomongoCollectionName)

			Expect(receivedErr).ToNot(HaveOccurred())
			Expect(receivedCollection).ToNot(BeNil())
		})
	})

	Context("when database is nil", func() {
		It("returns error", func() {
			receivedCollection, receivedErr := NewCollection[DummyDocument](nil, gomongoCollectionName)

			Expect(receivedErr).To(MatchError(ErrConnectionNotInitialized))
			Expect(receivedCollection).To(BeNil())
		})
	})

	Context("when database is not initialized", func() {
		It("returns error", func() {
			receivedCollection, receivedErr := NewCollection[DummyDocument](&Database{}, gomongoCollectionName)

			Expect(receivedErr).To(MatchError(ErrConnectionNotInitialized))
			Expect(receivedCollection).To(BeNil())
		})
	})

	Context("when mongo is down", func() {
		BeforeEach(func() {
			terminateMongoContainer(mongodbContainer, context.Background())
		})

		It("returns collection", func() {
			receivedCollection, receivedErr := NewCollection[DummyDocument](gomongoDatabase, gomongoCollectionName)

			Expect(receivedErr).ToNot(HaveOccurred())
			Expect(receivedCollection).ToNot(BeNil())
		})
	})
})
