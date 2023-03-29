package mongo

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ = Describe("Collection", func() {
	Describe("NewCollection", func() {
		It("returns correct collection instance", func() {
			expectedCollection := &collection{name: "Testing"}
			receivedCollection := NewCollection("Testing")

			Expect(receivedCollection).To(Equal(expectedCollection))
		})
	})

	Describe("init", func() {
		When("collection was not initialized", func() {
			BeforeEach(func() {
				client, _ := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
				mongoDatabase = client.Database("test_database")
			})

			AfterEach(func() {
				mongoDatabase = nil
			})
			It("returns nil and set mongoCollection", func() {
				c := collection{name: "test_init"}
				err := c.init()

				Expect(err).NotTo(HaveOccurred())
				Expect(c.mongoCollection).NotTo(BeNil())
			})
		})

		When("collection was already initialized", func() {
			It("returns nil", func() {
				initialMongoCollection := &mongo.Collection{}
				c := collection{mongoCollection: initialMongoCollection}
				err := c.init()

				Expect(err).NotTo(HaveOccurred())
				Expect(c.mongoCollection).To(Equal(initialMongoCollection))
			})
		})

		When("mongoDatabase was not initialized", func() {
			It("returns connection not inirialized error", func() {
				c := collection{}
				err := c.init()

				Expect(err).To(MatchError(ErrConnectionNotInitialized))
			})
		})
	})
	Describe("All", func() {
		// TODO
	})
	Describe("Create", func() {
		// TODO
	})
	Describe("Count", func() {
		// TODO
	})
	Describe("DeleteID", func() {
		// TODO
	})
	Describe("FindOne", func() {
		// TODO
	})
	Describe("DeleteID", func() {
		// TODO
	})
	Describe("FindOne", func() {
		// TODO
	})
	Describe("First", func() {
		// TODO
	})
	Describe("UpdateID", func() {
		// TODO
	})
	Describe("Where", func() {
		// TODO
	})
})
