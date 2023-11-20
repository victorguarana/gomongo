package gomongo

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
)

var _ = Describe("validate", func() {
	Describe("valid cases", func() {
		Context("when settings are valid", func() {
			It("returns nil", func() {
				cs := ConnectionSettings{
					URI:          "mongodb://localhost:27017/test",
					DatabaseName: "test",
					Timeout:      10 * time.Second,
				}

				receivedErr := cs.validate()

				Expect(receivedErr).NotTo(HaveOccurred())
			})
		})
	})

	Describe("invalid cases", func() {
		Context("when URI is empty", func() {
			It("returns ErrInvalidSettings", func() {
				cs := ConnectionSettings{
					DatabaseName: "test",
					Timeout:      10 * time.Second,
				}

				receivedErr := cs.validate()

				Expect(receivedErr).To(MatchError(ErrInvalidSettings))
				Expect(receivedErr).To(MatchError(ContainSubstring("URI is invalid")))
			})
		})

		Context("when DatabaseName is empty", func() {
			It("returns ErrInvalidSettings", func() {
				cs := ConnectionSettings{
					URI:     "mongodb://localhost:27017/test",
					Timeout: 10 * time.Second,
				}

				receivedErr := cs.validate()

				Expect(receivedErr).To(MatchError(ErrInvalidSettings))
				Expect(receivedErr).To(MatchError(ContainSubstring("DatabaseName is invalid")))
			})
		})

		Context("when Timeout is empty", func() {
			It("returns ErrInvalidSettings", func() {
				cs := ConnectionSettings{
					URI:          "mongodb://localhost:27017/test",
					DatabaseName: "test",
				}

				receivedErr := cs.validate()

				Expect(receivedErr).To(MatchError(ErrInvalidSettings))
				Expect(receivedErr).To(MatchError(ContainSubstring("Timeout is invalid")))
			})
		})

		Context("when Timeout is negative", func() {
			It("returns ErrInvalidSettings", func() {
				cs := ConnectionSettings{
					URI:          "mongodb://localhost:27017/test",
					DatabaseName: "test",
					Timeout:      -10 * time.Second,
				}

				receivedErr := cs.validate()

				Expect(receivedErr).To(MatchError(ErrInvalidSettings))
				Expect(receivedErr).To(MatchError(ContainSubstring("Timeout is invalid")))
			})
		})
	})

})

var _ = Describe("Init", func() {
	var (
		mongodbContainer    *mongodb.MongoDBContainer
		mongodbContainerURI string
	)

	BeforeEach(func() {
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

	Describe("success cases", func() {
		Context("when mongo is running", func() {
			It("returns nil", func() {
				cs := ConnectionSettings{
					URI:          mongodbContainerURI,
					DatabaseName: "test",
					Timeout:      10 * time.Second,
				}

				receivedErr := Init(cs)

				Expect(receivedErr).NotTo(HaveOccurred())
			})
		})
	})

	Describe("fail cases", func() {
		Context("when mongo is stopped", func() {
			BeforeEach(func() {
				if err := mongodbContainer.Stop(context.Background(), nil); err != nil {
					panic(err)
				}
			})

			AfterEach(func() {
				if err := mongodbContainer.Start(context.Background()); err != nil {
					panic(err)
				}
			})

			It("returns error", func() {
				cs := ConnectionSettings{
					URI:          mongodbContainerURI,
					DatabaseName: "test",
					Timeout:      1 * time.Second,
				}

				receivedErr := Init(cs)

				Expect(receivedErr).To(MatchError(ErrGomongoCanNotConnect))
			})
		})
	})
})
