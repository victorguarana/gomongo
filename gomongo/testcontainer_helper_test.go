package gomongo

import (
	"context"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"

	. "github.com/onsi/ginkgo/v2"
)

func runMongoContainer(ctx context.Context) (*mongodb.MongoDBContainer, string) {
	mongodbContainer, err := mongodb.RunContainer(ctx, testcontainers.WithImage("mongo:6"))
	if err != nil {
		panic(err)
	}

	mongodbContainerURI, err := mongodbContainer.ConnectionString(ctx)
	if err != nil {
		panic(err)
	}

	return mongodbContainer, mongodbContainerURI
}

func terminateMongoContainer(mongodbContainer *mongodb.MongoDBContainer, ctx context.Context) {
	if err := mongodbContainer.Terminate(ctx); err != nil {
		panic(err)
	}
}

func removeTestContainerLogs() {
	testcontainers.Logger = log.New(GinkgoWriter, "", log.LstdFlags)
}
