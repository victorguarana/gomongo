package mongo_test

// import (
// 	"log"
// 	"os"
// 	"path-builder/config"
// 	"path-builder/internal/database/mongo"
// 	"testing"

// 	. "github.com/onsi/ginkgo/v2"
// 	. "github.com/onsi/gomega"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// )

// var collectionName = "testCollection"

// type DataMock struct {
// 	ID    primitive.ObjectID `bson:"_id,omitempty"`
// 	Value int                `bson:"value,omitempty"`
// }

// func TestMongo(t *testing.T) {
// 	RegisterFailHandler(Fail)
// 	RunSpecs(t, "Mongo Suite")
// }

// var _ = BeforeSuite(func() {
// 	config.Set("DBAAS_MONGODB_ENDPOINT", os.Getenv("DBAAS_MONGODB_ENDPOINT"))
// 	Expect(config.Get("DBAAS_MONGODB_ENDPOINT")).NotTo(BeZero(), "You must set DBAAS_MONGODB_ENDPOINT environment variable with correct mongo address")
// })

// var _ = Describe("mongo", func() {
// 	Describe(".Init", func() {
// 		Context("when DBAAS_MONGODB_ENDPOINT env var is empty", func() {
// 			BeforeEach(func() {
// 				config.Set("DBAAS_MONGODB_ENDPOINT", "")
// 			})

// 			AfterEach(func() {
// 				config.Set("DBAAS_MONGODB_ENDPOINT", os.Getenv("DBAAS_MONGODB_ENDPOINT"))
// 			})

// 			It("panic and raise correct message", func() {
// 				Expect(func() { _ = mongo.Init("Test") }).To(PanicWith("you must set your 'DBAAS_MONGODB_ENDPOINT' environmental variable"))
// 			})
// 		})
// 	})

// 	Describe("#Create/#First", func() {
// 		var db *mongo.DB

// 		BeforeEach(func() {
// 			db = mongo.Init("Test")
// 		})

// 		Context("when failed to find first document", func() {
// 			It("return nil and correct error", func() {
// 				bson, err := db.First("wrongCollection")

// 				Expect(bson).To(BeNil())
// 				Expect(err).To(MatchError(mongo.ErrEmptyCollection))
// 			})
// 		})

// 		Context("when collection empty and one document is created", func() {
// 			var dataBSON bson.M

// 			BeforeEach(func() {
// 				var err error
// 				data := DataMock{
// 					Value: 5,
// 				}

// 				dataBSON, _ := mongo.DataToBSON(data)
// 				err = db.Create(collectionName, dataBSON)
// 				if err != nil {
// 					log.Fatal(err)
// 				}
// 			})

// 			AfterEach(func() {
// 				err := db.Delete(collectionName, dataBSON)
// 				if err != nil {
// 					log.Fatal(err)
// 				}
// 			})

// 			It("return first document correctly", func() {
// 				var err error
// 				var retrievedDataBSON bson.M
// 				retrievedDataBSON, err = db.First(collectionName)

// 				var retrievedData DataMock
// 				bsonBytes, _ := bson.Marshal(retrievedDataBSON)
// 				_ = bson.Unmarshal(bsonBytes, &retrievedData)

// 				Expect(retrievedData.Value).To(Equal(5))
// 				Expect(err).NotTo(HaveOccurred())
// 			})
// 		})
// 	})

// 	Describe("#Delete", func() {
// 		var db *mongo.DB

// 		BeforeEach(func() {
// 			db = mongo.Init("Test")
// 		})

// 		Context("when object to delete exists", func() {
// 			It("doesnt return error", func() {
// 				var err error
// 				data := DataMock{
// 					Value: 5,
// 				}

// 				dataBSON, _ := mongo.DataToBSON(data)
// 				err = db.Create(collectionName, dataBSON)
// 				if err != nil {
// 					log.Fatal(err)
// 				}

// 				err = db.Delete(collectionName, dataBSON)

// 				Expect(err).NotTo(HaveOccurred())
// 			})
// 		})

// 		Context("when object to delete not exists", func() {
// 			It("return error", func() {
// 				dataToDelete, _ := mongo.DataToBSON(DataMock{
// 					Value: 3,
// 				})

// 				err := db.Delete(collectionName, dataToDelete)

// 				Expect(err).To(MatchError(mongo.ErrNothingDeleted))
// 			})
// 		})
// 	})
// })
