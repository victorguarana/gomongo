package gomongo

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/bson/primitive"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type DummyStruct struct {
	ID           ID `bson:"_id" custom:"-"`
	Int          int
	Int8         int8
	Int16        int16
	Int32        int32
	Int64        int64
	String       string
	Bool         bool
	SString      []string
	SInt         []int
	SInt8        []int8
	SInt16       []int16
	SInt32       []int32
	SInt64       []int64
	SFloat32     []float32
	SFloat64     []float64
	SBool        []bool
	NestedStruct DummyNestedStruct
}

type DummyNestedStruct struct {
	Int                 int
	SecondInt           int
	AnotherNestedStruct DummySecondNestedStruct
}

type DummySecondNestedStruct struct {
	String string
}

var _ = Describe("NewCollection", Ordered, func() {
	var (
		databaseName   = "database_test"
		collectionName = "collection_test"

		mongodbContainerURI string
		mongodbContainer    *mongodb.MongoDBContainer

		gomongoDatabase *Database
	)

	BeforeAll(func() {
		mongodbContainer, mongodbContainerURI = runMongoContainer(context.Background())
		gomongoDatabase, _ = NewDatabase(ConnectionSettings{
			URI:               mongodbContainerURI,
			DatabaseName:      databaseName,
			ConnectionTimeout: time.Second,
		})
	})

	Context("when database is initialized", func() {
		It("returns collection", func() {
			receivedCollection, receivedErr := NewCollection[DummyStruct](gomongoDatabase, collectionName)

			Expect(receivedErr).ToNot(HaveOccurred())
			Expect(receivedCollection).ToNot(BeNil())
		})
	})

	Context("when database is nil", func() {
		It("returns error", func() {
			receivedCollection, receivedErr := NewCollection[DummyStruct](nil, collectionName)

			Expect(receivedErr).To(MatchError(ErrConnectionNotInitialized))
			Expect(receivedCollection).To(BeNil())
		})
	})

	Context("when database is not initialized", func() {
		It("returns error", func() {
			receivedCollection, receivedErr := NewCollection[DummyStruct](&Database{}, collectionName)

			Expect(receivedErr).To(MatchError(ErrConnectionNotInitialized))
			Expect(receivedCollection).To(BeNil())
		})
	})

	Context("when mongo is down", func() {
		BeforeEach(func() {
			terminateMongoContainer(mongodbContainer, context.Background())
		})

		It("returns collection", func() {
			receivedCollection, receivedErr := NewCollection[DummyStruct](gomongoDatabase, collectionName)

			Expect(receivedErr).ToNot(HaveOccurred())
			Expect(receivedCollection).ToNot(BeNil())
		})
	})
})

var _ = Describe("collection{}", Ordered, func() {
	var (
		databaseName   = "database_test"
		collectionName = "collection_test"

		mongodbContainerURI string
		mongodbContainer    *mongodb.MongoDBContainer

		sut collection[DummyStruct]
		err error
	)

	BeforeAll(func() {
		mongodbContainer, mongodbContainerURI = runMongoContainer(context.Background())
		sut, err = initializeCollection(mongodbContainerURI, databaseName, collectionName)
		if err != nil {
			Fail(err.Error())
		}
	})

	AfterAll(func() {
		terminateMongoContainer(mongodbContainer, context.Background())
	})

	Describe("All", Ordered, func() {
		AfterAll(func() {
			if err := sut.Drop(context.Background()); err != nil {
				Fail(err.Error())
			}
		})

		Context("when collection is empty", func() {
			It("should return empty slice and no error", func() {
				receivedStructs, receivedErr := sut.All(context.Background())

				Expect(receivedErr).ToNot(HaveOccurred())
				Expect(receivedStructs).To(BeEmpty())
			})
		})

		Context("when collection is not empty", func() {
			var (
				dummyStructs []DummyStruct
			)

			BeforeAll(func() {
				By("populating with Create")
				documentCount := randomIntBetween(10, 20)
				dummyStructs, err = populateCollectionWithManyFakeDocuments(sut, documentCount)
				if err != nil {
					Fail(err.Error())
				}
			})

			It("should return all structs and no error", func() {
				receivedStructs, receivedErr := sut.All(context.Background())

				Expect(receivedErr).ToNot(HaveOccurred())
				Expect(receivedStructs).To(Equal(dummyStructs))
			})
		})
	})

	Describe("Count", Ordered, func() {
		AfterAll(func() {
			if err := sut.Drop(context.Background()); err != nil {
				Fail(err.Error())
			}
		})

		Context("when collection is empty", func() {
			It("should return 0 and no error", func() {
				receivedCount, receivedErr := sut.Count(context.Background())

				Expect(receivedErr).ToNot(HaveOccurred())
				Expect(receivedCount).To(Equal(0))
			})
		})

		Context("when collection is not empty", func() {
			var expectedCount int

			BeforeAll(func() {
				By("populating with Create")
				expectedCount = randomIntBetween(10, 20)
				_, err := populateCollectionWithManyFakeDocuments(sut, expectedCount)
				if err != nil {
					Fail(err.Error())
				}
			})

			It("should return count and no error", func() {
				receivedCount, receivedErr := sut.Count(context.Background())

				Expect(receivedErr).ToNot(HaveOccurred())
				Expect(receivedCount).To(Equal(expectedCount))
			})
		})
	})

	Describe("DeleteID", Ordered, func() {
		var (
			deleteID ID
		)

		AfterAll(func() {
			if err := sut.Drop(context.Background()); err != nil {
				Fail(err.Error())
			}
		})

		Context("when collection ID is nil", func() {
			It("should return error", func() {
				receivedErr := sut.DeleteID(context.Background(), deleteID)

				Expect(receivedErr).To(MatchError(ErrEmptyID))
			})
		})

		Context("when collection is empty", func() {
			BeforeAll(func() {
				deleteID, err = notExistentID()
				if err != nil {
					Fail(err.Error())
				}
			})

			It("should return error", func() {
				receivedErr := sut.DeleteID(context.Background(), deleteID)

				Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
			})
		})

		Context("when collection is not empty", func() {
			var (
				dummyStructs []DummyStruct
			)

			BeforeAll(func() {
				By("populating with Create")
				var err error
				documentCount := randomIntBetween(10, 20)
				dummyStructs, err = populateCollectionWithManyFakeDocuments(sut, documentCount)
				if err != nil {
					Fail(err.Error())
				}
			})

			Context("when ID does not exist", func() {
				BeforeAll(func() {
					deleteID, err = notExistentID()
					if err != nil {
						Fail(err.Error())
					}
				})

				It("should return error", func() {
					receivedErr := sut.DeleteID(context.Background(), deleteID)

					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
				})

				It("should not change document count", func() {
					By("validating with Count")
					receivedCount, receivedErr := sut.Count(context.Background())

					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedCount).To(Equal(len(dummyStructs)))
				})
			})

			Context("when ID is from first document", func() {
				BeforeAll(func() {
					first := dummyStructs[0]
					deleteID = first.ID
				})

				AfterAll(func() {
					dummyStructs = dummyStructs[1:]
				})

				It("should return no error", func() {
					receivedErr := sut.DeleteID(context.Background(), deleteID)

					Expect(receivedErr).NotTo(HaveOccurred())
				})

				It("should not find deleted document", func() {
					By("validating with FindID")
					_, receivedErr := sut.FindID(context.Background(), deleteID)

					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
				})

				It("should have one less document", func() {
					By("validating with Count")
					receivedCount, receivedErr := sut.Count(context.Background())

					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedCount).To(Equal(len(dummyStructs) - 1))
				})
			})

			Context("when ID is from last document", func() {
				BeforeAll(func() {
					last := dummyStructs[len(dummyStructs)-1]
					deleteID = last.ID
				})

				AfterAll(func() {
					dummyStructs = dummyStructs[:len(dummyStructs)-1]
				})

				It("should return no error", func() {
					receivedErr := sut.DeleteID(context.Background(), deleteID)

					Expect(receivedErr).NotTo(HaveOccurred())
				})

				It("should not find deleted document", func() {
					By("validating with FindID")
					_, receivedErr := sut.FindID(context.Background(), deleteID)

					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
				})

				It("should have one less document", func() {
					By("validating with Count")
					receivedCount, receivedErr := sut.Count(context.Background())

					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedCount).To(Equal(len(dummyStructs) - 1))
				})
			})

			Context("when ID is in the middle of the collection", func() {
				BeforeAll(func() {
					middleDummy := dummyStructs[len(dummyStructs)/2]
					deleteID = middleDummy.ID
				})

				AfterAll(func() {
					dummyStructs = append(
						dummyStructs[:len(dummyStructs)/2-1],
						dummyStructs[len(dummyStructs)/2+1:]...)
				})

				It("should return no error", func() {
					receivedErr := sut.DeleteID(context.Background(), deleteID)

					Expect(receivedErr).NotTo(HaveOccurred())
				})

				It("should not find deleted document", func() {
					By("validating with FindID")
					_, receivedErr := sut.FindID(context.Background(), deleteID)

					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
				})

				It("should have one less document", func() {
					By("validating with Count")
					receivedCount, receivedErr := sut.Count(context.Background())

					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedCount).To(Equal(len(dummyStructs) - 1))
				})
			})
		})
	})

	Describe("Create", Ordered, func() {
		var (
			dummy             DummyStruct
			initialID         ID
			receivedCreateErr error
		)

		BeforeAll(func() {
			if err := fakeData(&dummy); err != nil {
				Fail(err.Error())
			}

			var err error
			initialID, err = notExistentID()
			if err != nil {
				Fail(err.Error())
			}

			dummy.ID = initialID
			dummy.ID, receivedCreateErr = sut.Create(context.Background(), dummy)
		})

		AfterAll(func() {
			if err := sut.Drop(context.Background()); err != nil {
				Fail(err.Error())
			}
		})

		It("should override and return valid ID and no error", func() {
			Expect(receivedCreateErr).ToNot(HaveOccurred())
			Expect(dummy.ID).ToNot(BeNil())
			Expect(dummy.ID).ToNot(Equal(initialID))
		})

		It("should insert document with correct fields", func() {
			By("validating with First")
			receivedDummy, receivedErr := sut.First(context.Background())

			Expect(receivedErr).ToNot(HaveOccurred())
			Expect(receivedDummy).To(Equal(dummy))
		})

		It("should insert just one document", func() {
			By("validating with Count")
			receivedCount, receivedErr := sut.Count(context.Background())

			Expect(receivedErr).ToNot(HaveOccurred())
			Expect(receivedCount).To(Equal(1))
		})
	})

	Describe("Drop", Ordered, func() {
		Context("when collection is empty", func() {
			It("should return no error", func() {
				receivedErr := sut.Drop(context.Background())

				Expect(receivedErr).ToNot(HaveOccurred())
			})
		})

		Context("when collection in not empty", func() {
			BeforeAll(func() {
				documentCount := randomIntBetween(10, 20)
				_, err := populateCollectionWithManyFakeDocuments(sut, documentCount)
				if err != nil {
					Fail(err.Error())
				}
			})

			It("should return no error", func() {
				receivedErr := sut.Drop(context.Background())

				Expect(receivedErr).ToNot(HaveOccurred())
			})

			It("should have dropped all documents", func() {
				By("validating with Count")
				receivedCount, receivedErr := sut.Count(context.Background())

				Expect(receivedErr).ToNot(HaveOccurred())
				Expect(receivedCount).To(Equal(0))
			})
		})
	})

	Describe("FindID", func() {
		var (
			findID ID
		)

		AfterAll(func() {
			if err := sut.Drop(context.Background()); err != nil {
				Fail(err.Error())
			}
		})

		Context("when id is nil", func() {
			It("shold return document not found error", func() {
				receivedDummy, receivedErr := sut.FindID(context.Background(), nil)

				Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
				Expect(receivedDummy).To(Equal(DummyStruct{}))
			})
		})

		Context("when collection is empty", func() {
			BeforeAll(func() {
				var err error
				findID, err = notExistentID()
				if err != nil {
					Fail(err.Error())
				}
			})

			It("should return document not found error", func() {
				receivedDummy, receivedErr := sut.FindID(context.Background(), findID)

				Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
				Expect(receivedDummy).To(Equal(DummyStruct{}))
			})
		})

		Context("when collection is not empty", func() {
			var (
				dummyStructs  []DummyStruct
				documentCount int
				receivedDummy DummyStruct
				receivedErr   error
				expectedDummy DummyStruct
			)

			BeforeAll(func() {
				documentCount = randomIntBetween(10, 20)
				dummyStructs, err = populateCollectionWithManyFakeDocuments(sut, documentCount)
				if err != nil {
					Fail(err.Error())
				}
			})

			Context("when ID does not exist", func() {
				BeforeAll(func() {
					_, err := notExistentID()
					if err != nil {
						Fail(err.Error())
					}
				})

				It("should return document not found error", func() {
					receivedDummy, receivedErr = sut.FindID(context.Background(), nil)

					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
					Expect(receivedDummy).To(Equal(DummyStruct{}))
				})
			})

			Context("when ID is from first document", func() {
				BeforeAll(func() {
					By("getting ID with First")
					expectedDummy = dummyStructs[0]
					receivedDummy, receivedErr = sut.FindID(context.Background(), expectedDummy.ID)
				})

				It("should return no error", func() {
					Expect(receivedErr).ToNot(HaveOccurred())
				})

				It("should return correct document", func() {
					Expect(receivedDummy).To(Equal(dummyStructs[0]))
				})
			})

			Context("when ID is from last document", func() {
				BeforeAll(func() {
					By("getting ID with Last")
					expectedDummy = dummyStructs[len(dummyStructs)-1]
					receivedDummy, receivedErr = sut.FindID(context.Background(), expectedDummy.ID)
				})

				It("should return no error", func() {
					Expect(receivedErr).ToNot(HaveOccurred())
				})

				It("should return correct document", func() {
					Expect(receivedDummy).To(Equal(expectedDummy))
				})
			})

			Context("when ID is from the middle of the collection", func() {
				BeforeAll(func() {
					By("getting ID with Last")
					expectedDummy = dummyStructs[len(dummyStructs)/2]
					receivedDummy, receivedErr = sut.FindID(context.Background(), expectedDummy.ID)
				})

				It("should return no error", func() {
					Expect(receivedErr).ToNot(HaveOccurred())
				})

				It("should return correct document", func() {
					Expect(receivedDummy).To(Equal(expectedDummy))
				})
			})
		})
	})
})

func initializeCollection(mongoURI, databaseName, collectionName string) (collection[DummyStruct], error) {
	gomongoDatabase, err := NewDatabase(ConnectionSettings{
		URI:               mongoURI,
		DatabaseName:      databaseName,
		ConnectionTimeout: time.Second,
	})

	if err != nil {
		return collection[DummyStruct]{}, fmt.Errorf("Could not create database: %e", err)
	}

	sut := collection[DummyStruct]{
		mongoCollection: gomongoDatabase.mongoDatabase.Collection(collectionName),
	}

	return sut, nil
}

func fakeData(dummy *DummyStruct) error {
	if err := faker.FakeData(dummy); err != nil {
		return fmt.Errorf("Could not generate fake data: %e", err)
	}

	return nil
}

func randomIntBetween(min, max int) int {
	return rand.Intn(max-min) + min
}

func populateCollectionWithManyFakeDocuments(collection collection[DummyStruct], n int) ([]DummyStruct, error) {
	dummyStructs, err := generateDummyStructs(n)
	if err != nil {
		return nil, err
	}

	if err := insertManyInCollection(collection, dummyStructs); err != nil {
		return nil, err
	}

	return dummyStructs, nil
}

func generateDummyStructs(n int) ([]DummyStruct, error) {
	dummyStructs := make([]DummyStruct, n)
	for i := range dummyStructs {
		if err := fakeData(&dummyStructs[i]); err != nil {
			return nil, err
		}
	}

	return dummyStructs, nil
}

func insertManyInCollection(collection collection[DummyStruct], dummyStructs []DummyStruct) error {
	for i, dummy := range dummyStructs {
		var err error
		dummyStructs[i].ID, err = collection.Create(context.Background(), dummy)
		if err != nil {
			return fmt.Errorf("Could not populate collection: %e", err)
		}
	}

	return nil
}
func notExistentID() (ID, error) {
	objectID, err := primitive.ObjectIDFromHex("60f3b3b3b3b3b3b3b3b3b3b3")
	if err != nil {
		return nil, err
	}
	return ID(&objectID), nil
}
