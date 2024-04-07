package gomongo

import (
	"context"
	"log"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
)

var _ = Describe("NewDatabase", func() {
	var (
		mongodbContainer    *mongodb.MongoDBContainer
		mongodbContainerURI string
	)

	Describe("success cases", func() {
		BeforeEach(func() {
			testcontainers.Logger = log.New(GinkgoWriter, "", log.LstdFlags)

			var err error
			mongodbContainer, err = mongodb.RunContainer(context.Background(), testcontainers.WithImage("mongo:6"))
			if err != nil {
				panic(err)
			}

			mongodbContainerURI, err = mongodbContainer.ConnectionString(context.Background())
			if err != nil {
				panic(err)
			}
		})

		AfterEach(func() {
			if err := mongodbContainer.Terminate(context.Background()); err != nil {
				panic(err)
			}
		})

		Context("when mongo is running", func() {
			It("returns nil", func() {
				cs := ConnectionSettings{
					URI:          mongodbContainerURI,
					DatabaseName: "test",
				}

				receivedDatabase, receivedErr := NewDatabase(cs)

				Expect(receivedErr).NotTo(HaveOccurred())
				Expect(receivedDatabase.mongoDatabse).NotTo(BeNil())
			})
		})
	})

	Describe("fail cases", func() {
		Context("when mongo was not started", func() {
			It("returns error", func() {
				cs := ConnectionSettings{
					URI:               "mongogo://localhost:27017",
					DatabaseName:      "test",
					ConnectionTimeout: 1 * time.Second,
				}

				receivedDatabase, receivedErr := NewDatabase(cs)

				Expect(receivedErr).To(MatchError(ErrGomongoCanNotConnect))
				Expect(receivedDatabase).To(BeNil())
			})
		})

		Context("when mongo is stopped", func() {
			BeforeEach(func() {
				testcontainers.Logger = log.New(GinkgoWriter, "", log.LstdFlags)

				var err error
				mongodbContainer, err = mongodb.RunContainer(context.Background(), testcontainers.WithImage("mongo:6"))
				if err != nil {
					panic(err)
				}

				mongodbContainerURI, err = mongodbContainer.ConnectionString(context.Background())
				if err != nil {
					panic(err)
				}

				if err := mongodbContainer.Stop(context.Background(), nil); err != nil {
					panic(err)
				}
			})

			AfterEach(func() {
				if err := mongodbContainer.Terminate(context.Background()); err != nil {
					panic(err)
				}
			})

			It("returns error", func() {
				cs := ConnectionSettings{
					URI:               mongodbContainerURI,
					DatabaseName:      "test",
					ConnectionTimeout: 1 * time.Second,
				}

				receivedDatabase, receivedErr := NewDatabase(cs)

				Expect(receivedErr).To(MatchError(ErrGomongoCanNotConnect))
				Expect(receivedDatabase).To(BeNil())
			})
		})
	})
})
