package gomongo

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	. "github.com/onsi/ginkgo/v2"
)

func runMongoContainer(ctx context.Context) (testcontainers.Container, string) {
	req := testcontainers.ContainerRequest{
		Image:        "mongo:4.2",
		Cmd:          []string{"--replSet", "local", "--bind_ip_all"},
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForLog("waiting for connections on port 27017"),
	}

	mongoC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}

	_, _, err = mongoC.Exec(ctx, []string{"mongo", "--eval", "rs.initiate({_id: \"local\", members: [{ _id : 0, host : \"localhost:27017\"}] })"})
	if err != nil {
		panic(err)
	}

	mongoC.Endpoint(ctx, "mongodb")

	mongodbContainerURI, err := mongoC.Endpoint(ctx, "mongodb")
	if err != nil {
		panic(err)
	}

	mongodbContainerURI = fmt.Sprintf("%s/test?replicaSet=local", mongodbContainerURI)

	return mongoC, mongodbContainerURI
}

func terminateContainer(mongodbContainer testcontainers.Container, ctx context.Context) {
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
