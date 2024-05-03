package gomongo_test

import (
	"context"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/victorguarana/gomongo"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NewDatabase", Ordered, func() {
	var (
		mongodbContainer    *mongodb.MongoDBContainer
		mongodbContainerURI string
		connectionSettings  gomongo.ConnectionSettings
	)

	BeforeAll(func() {
		mongodbContainer, mongodbContainerURI = runMongoContainer(context.Background())
	})

	Context("when connection settings param is invalid", func() {
		Context("when URI is empty", func() {
			It("returns ErrInvalidSettings", func() {
				connectionSettings := gomongo.ConnectionSettings{
					DatabaseName: "test",
				}

				receivedDatabase, receivedErr := gomongo.NewDatabase(context.Background(), connectionSettings)
				Expect(receivedErr).To(MatchError(gomongo.ErrInvalidSettings))
				Expect(receivedErr).To(MatchError(ContainSubstring("URI can not be empty")))
				Expect(receivedDatabase).To(Equal(gomongo.Database{}))
			})
		})

		Context("when DatabaseName is empty", func() {
			It("returns ErrInvalidSettings", func() {
				connectionSettings := gomongo.ConnectionSettings{
					URI: mongodbContainerURI,
				}

				receivedDatabase, receivedErr := gomongo.NewDatabase(context.Background(), connectionSettings)
				Expect(receivedErr).To(MatchError(gomongo.ErrInvalidSettings))
				Expect(receivedErr).To(MatchError(ContainSubstring("Database Name can not be empty")))
				Expect(receivedDatabase).To(Equal(gomongo.Database{}))
			})
		})
	})

	Context("when connection settings param is valid", func() {
		Context("when mongo is up", func() {
			Context("when all settings are filled", func() {
				It("returns nil", func() {
					connectionSettings := gomongo.ConnectionSettings{
						URI:               mongodbContainerURI,
						DatabaseName:      "test",
						ConnectionTimeout: 10 * time.Second,
					}

					receivedDatabase, receivedErr := gomongo.NewDatabase(context.Background(), connectionSettings)
					Expect(receivedErr).NotTo(HaveOccurred())
					Expect(receivedDatabase).NotTo(Equal(gomongo.Database{}))
				})
			})

			Context("when timeout is empty", func() {
				It("returns nil", func() {
					connectionSettings := gomongo.ConnectionSettings{
						URI:          mongodbContainerURI,
						DatabaseName: "test",
					}

					receivedDatabase, receivedErr := gomongo.NewDatabase(context.Background(), connectionSettings)
					Expect(receivedErr).NotTo(HaveOccurred())
					Expect(receivedDatabase).NotTo(Equal(gomongo.Database{}))
				})
			})
		})

		Context("when mongo is down", func() {
			BeforeAll(func() {
				terminateMongoContainer(mongodbContainer, context.Background())
			})

			It("returns error", func() {
				connectionSettings = gomongo.ConnectionSettings{
					URI:               mongodbContainerURI,
					DatabaseName:      "test",
					ConnectionTimeout: 1 * time.Second,
				}

				receivedDatabase, receivedErr := gomongo.NewDatabase(context.Background(), connectionSettings)
				Expect(receivedErr).To(MatchError(gomongo.ErrGomongoCanNotConnect))
				Expect(receivedDatabase).To(Equal(gomongo.Database{}))
			})
		})
	})
})
