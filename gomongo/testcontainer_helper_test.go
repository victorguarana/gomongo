package gomongo

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"

	. "github.com/onsi/ginkgo/v2"
)

func runMongoContainer(ctx context.Context) (*mongodb.MongoDBContainer, string) {
	mongodbContainer, err := mongodb.RunContainer(ctx, testcontainers.WithImage(getMongoImageName()))
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

func getMongoImageName() string {
	versionFromEnv := os.Getenv("MONGO_VERSION")
	return fmt.Sprintf("mongo:%s", versionFromEnv)
}
