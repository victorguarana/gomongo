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
		It("should return collection", func() {
			receivedCollection, receivedErr := NewCollection[DummyStruct](gomongoDatabase, collectionName)
			Expect(receivedErr).ToNot(HaveOccurred())
			Expect(receivedCollection).ToNot(BeNil())
		})
	})

	Context("when database is nil", func() {
		It("should return error", func() {
			receivedCollection, receivedErr := NewCollection[DummyStruct](nil, collectionName)
			Expect(receivedErr).To(MatchError(ErrConnectionNotInitialized))
			Expect(receivedCollection).To(BeNil())
		})
	})

	Context("when database is not initialized", func() {
		It("should return error", func() {
			receivedCollection, receivedErr := NewCollection[DummyStruct](&Database{}, collectionName)
			Expect(receivedErr).To(MatchError(ErrConnectionNotInitialized))
			Expect(receivedCollection).To(BeNil())
		})
	})

	Context("when mongo is down", func() {
		BeforeEach(func() {
			terminateMongoContainer(mongodbContainer, context.Background())
		})

		It("should return collection", func() {
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
		Context("when collection is empty", func() {
			It("should return empty slice and no error", func() {
				receivedStructs, receivedErr := sut.All(context.Background())
				Expect(receivedErr).ToNot(HaveOccurred())
				Expect(receivedStructs).To(BeEmpty())
			})
		})

		Context("when collection is filled", func() {
			var dummies []DummyStruct

			BeforeAll(func() {
				By("populating with Create")
				documentCount := randomIntBetween(10, 20)
				dummies, err = populateCollectionWithManyFakeDocuments(sut, documentCount)
				if err != nil {
					Fail(err.Error())
				}
			})

			AfterAll(func() {
				if err := sut.Drop(context.Background()); err != nil {
					Fail(err.Error())
				}
			})

			It("should return all structs and no error", func() {
				receivedStructs, receivedErr := sut.All(context.Background())
				Expect(receivedErr).ToNot(HaveOccurred())
				Expect(receivedStructs).To(Equal(dummies))
			})
		})
	})

	Describe("Count", Ordered, func() {
		Context("when collection is empty", func() {
			It("should return 0 and no error", func() {
				receivedCount, receivedErr := sut.Count(context.Background())
				Expect(receivedErr).ToNot(HaveOccurred())
				Expect(receivedCount).To(Equal(0))
			})
		})

		Context("when collection is filled", func() {
			var expectedCount int

			BeforeAll(func() {
				By("populating with Create")
				expectedCount = randomIntBetween(10, 20)
				_, err := populateCollectionWithManyFakeDocuments(sut, expectedCount)
				if err != nil {
					Fail(err.Error())
				}
			})

			AfterAll(func() {
				if err := sut.Drop(context.Background()); err != nil {
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

	Describe("DeleteID", func() {
		Context("when collection ID is nil", func() {
			It("should return error", func() {
				receivedErr := sut.DeleteID(context.Background(), nil)
				Expect(receivedErr).To(MatchError(ErrEmptyID))
			})
		})

		Context("when collection is empty", func() {
			var deleteID ID

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

		Context("when collection is filled", func() {
			var (
				deleteID ID
				dummies  []DummyStruct
			)

			BeforeAll(func() {
				By("populating with Create")
				var err error
				documentCount := randomIntBetween(10, 20)
				dummies, err = populateCollectionWithManyFakeDocuments(sut, documentCount)
				if err != nil {
					Fail(err.Error())
				}
			})

			AfterAll(func() {
				if err := sut.Drop(context.Background()); err != nil {
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

				It("should return error and not delete any document", func() {
					receivedErr := sut.DeleteID(context.Background(), deleteID)
					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))

					By("validating with All")
					receivedStructs, receivedErr := sut.All(context.Background())
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(Equal(dummies))
				})
			})

			Context("when ID is from first document", func() {
				BeforeAll(func() {
					first := dummies[0]
					deleteID = first.ID
					dummies = dummies[1:]
				})

				It("should return no error and delete document", func() {
					receivedErr := sut.DeleteID(context.Background(), deleteID)
					Expect(receivedErr).NotTo(HaveOccurred())

					By("validating with All")
					receivedStructs, receivedErr := sut.All(context.Background())
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(Equal(dummies))
				})
			})

			Context("when ID is from last document", func() {
				BeforeAll(func() {
					last := dummies[len(dummies)-1]
					deleteID = last.ID
					dummies = dummies[:len(dummies)-1]
				})

				It("should return no error and delete document", func() {
					receivedErr := sut.DeleteID(context.Background(), deleteID)
					Expect(receivedErr).NotTo(HaveOccurred())

					By("validating with All")
					receivedStructs, receivedErr := sut.All(context.Background())
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(Equal(dummies))
				})
			})

			Context("when ID is in the middle of the collection", func() {
				BeforeAll(func() {
					middleDummy := dummies[len(dummies)/2]
					deleteID = middleDummy.ID
					dummies = append(
						dummies[:len(dummies)/2],
						dummies[len(dummies)/2+1:]...)
				})

				It("should return no error and delete document", func() {
					receivedErr := sut.DeleteID(context.Background(), deleteID)
					Expect(receivedErr).NotTo(HaveOccurred())

					By("validating with All")
					receivedStructs, receivedErr := sut.All(context.Background())
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(Equal(dummies))
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

		})

		AfterAll(func() {
			if err := sut.Drop(context.Background()); err != nil {
				Fail(err.Error())
			}
		})

		It("should override existent id, return valid ID and no error, and insert just one document with correct fields", func() {
			dummy.ID, receivedCreateErr = sut.Create(context.Background(), dummy)
			Expect(receivedCreateErr).ToNot(HaveOccurred())
			Expect(dummy.ID).ToNot(BeNil())
			Expect(dummy.ID).ToNot(Equal(initialID))

			By("validating with All")
			receivedStructs, receivedErr := sut.All(context.Background())
			Expect(receivedErr).ToNot(HaveOccurred())
			Expect(receivedStructs).To(Equal([]DummyStruct{dummy}))
		})
	})

	Describe("Drop", Ordered, func() {
		Context("when collection is empty", func() {
			It("should return no error", func() {
				receivedErr := sut.Drop(context.Background())
				Expect(receivedErr).ToNot(HaveOccurred())
			})
		})

		Context("when collection in filled", func() {
			BeforeAll(func() {
				By("populating with Create")
				documentCount := randomIntBetween(10, 20)
				_, err := populateCollectionWithManyFakeDocuments(sut, documentCount)
				if err != nil {
					Fail(err.Error())
				}
			})

			It("should return no error and drop all documents", func() {
				receivedErr := sut.Drop(context.Background())
				Expect(receivedErr).ToNot(HaveOccurred())

				By("validating with Count")
				receivedCount, receivedErr := sut.Count(context.Background())
				Expect(receivedErr).ToNot(HaveOccurred())
				Expect(receivedCount).To(Equal(0))
			})
		})
	})

	Describe("FindID", func() {
		var findID ID

		Context("when id is nil", func() {
			It("should return empty id error", func() {
				receivedDummy, receivedErr := sut.FindID(context.Background(), nil)
				Expect(receivedErr).To(MatchError(ErrEmptyID))
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

		Context("when collection is filled", func() {
			var (
				dummies       []DummyStruct
				expectedDummy DummyStruct
			)

			BeforeAll(func() {
				By("populating with Create")
				documentCount := randomIntBetween(10, 20)
				dummies, err = populateCollectionWithManyFakeDocuments(sut, documentCount)
				if err != nil {
					Fail(err.Error())
				}
			})

			AfterAll(func() {
				if err := sut.Drop(context.Background()); err != nil {
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

				It("should return empty id error", func() {
					receivedDummy, receivedErr := sut.FindID(context.Background(), nil)
					Expect(receivedErr).To(MatchError(ErrEmptyID))
					Expect(receivedDummy).To(Equal(DummyStruct{}))
				})
			})

			Context("when ID is from first document", func() {
				BeforeAll(func() {
					By("getting ID with First")
					expectedDummy = dummies[0]
				})

				It("should return correct document and no error", func() {
					receivedDummy, receivedErr := sut.FindID(context.Background(), expectedDummy.ID)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedDummy).To(Equal(dummies[0]))
				})
			})

			Context("when ID is from last document", func() {
				BeforeAll(func() {
					By("getting ID with Last")
					expectedDummy = dummies[len(dummies)-1]
				})

				It("should return correct document and no error", func() {
					receivedDummy, receivedErr := sut.FindID(context.Background(), expectedDummy.ID)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedDummy).To(Equal(expectedDummy))
				})
			})

			Context("when ID is from the middle of the collection", func() {
				BeforeAll(func() {
					By("getting ID with Last")
					expectedDummy = dummies[len(dummies)/2]
				})

				It("should return correct document and no error", func() {
					receivedDummy, receivedErr := sut.FindID(context.Background(), expectedDummy.ID)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedDummy).To(Equal(expectedDummy))
				})
			})
		})
	})

	Describe("FindOne", func() {
		var filter any

		Context("when collection is empty", func() {
			It("should return document not found error", func() {
				receivedDummy, receivedErr := sut.FindOne(context.Background(), nil)
				Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
				Expect(receivedDummy).To(Equal(DummyStruct{}))
			})
		})

		Context("when collection is filled", func() {
			var (
				dummies       []DummyStruct
				expectedDummy DummyStruct
			)

			BeforeAll(func() {
				By("populating with Create")
				documentCount := randomIntBetween(10, 20)
				dummies, err = populateCollectionWithManyFakeDocuments(sut, documentCount)
				if err != nil {
					Fail(err.Error())
				}
			})

			AfterAll(func() {
				if err := sut.Drop(context.Background()); err != nil {
					Fail(err.Error())
				}
			})

			Context("when filter is nil", func() {
				BeforeAll(func() {
					expectedDummy = dummies[0]
				})

				It("should return first document and no error", func() {
					receivedDummy, receivedErr := sut.FindOne(context.Background(), nil)
					Expect(receivedErr).NotTo(HaveOccurred())
					Expect(receivedDummy).To(Equal(expectedDummy))
				})
			})

			Context("when filter is empty", func() {
				BeforeAll(func() {
					expectedDummy = dummies[0]
					filter = map[string]any{}
				})

				It("should return first document and no error", func() {
					receivedDummy, receivedErr := sut.FindOne(context.Background(), filter)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedDummy).To(Equal(expectedDummy))
				})
			})

			Context("when filter has non existent fields", func() {
				BeforeAll(func() {
					filter = map[string]any{"nonexistent": 0}
				})

				It("should return document not found error", func() {
					receivedDummy, receivedErr := sut.FindOne(context.Background(), filter)
					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
					Expect(receivedDummy).To(Equal(DummyStruct{}))
				})
			})

			Context("when filter has wrong type values", func() {
				BeforeAll(func() {
					filter = map[string]any{"string": -1}
				})

				It("should return document not found error", func() {
					receivedDummy, receivedErr := sut.FindOne(context.Background(), filter)
					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
					Expect(receivedDummy).To(Equal(DummyStruct{}))
				})
			})

			Context("when filter does not match any document", func() {
				BeforeAll(func() {
					filter = map[string]any{"string": ""}
				})

				It("should return document not found error", func() {
					receivedDummy, receivedErr := sut.FindOne(context.Background(), filter)
					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
					Expect(receivedDummy).To(Equal(DummyStruct{}))
				})
			})

			Context("when filter matches a document", func() {
				BeforeAll(func() {
					expectedDummy = dummies[len(dummies)/2]
					filter = map[string]any{"string": expectedDummy.String}
				})

				It("should return correct document and no error", func() {
					receivedDummy, receivedErr := sut.FindOne(context.Background(), filter)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedDummy).To(Equal(expectedDummy))
				})
			})
		})
	})

	Describe("First", func() {
		Context("when collection is empty", func() {
			It("should return document not found error", func() {
				receivedDummy, receivedErr := sut.First(context.Background())
				Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
				Expect(receivedDummy).To(Equal(DummyStruct{}))
			})
		})

		Context("when collection is filled", func() {
			var expectedDummy DummyStruct

			BeforeAll(func() {
				By("populating with Create")
				documentCount := randomIntBetween(10, 20)
				dummies, err := populateCollectionWithManyFakeDocuments(sut, documentCount)
				if err != nil {
					Fail(err.Error())
				}

				expectedDummy = dummies[0]
			})

			AfterAll(func() {
				if err := sut.Drop(context.Background()); err != nil {
					Fail(err.Error())
				}
			})

			It("should return first document and no error", func() {
				receivedDummy, receivedErr := sut.First(context.Background())
				Expect(receivedErr).ToNot(HaveOccurred())
				Expect(receivedDummy).To(Equal(expectedDummy))
			})
		})
	})

	Describe("FirstInserted", func() {
		var filter any

		AfterAll(func() {
			if err := sut.Drop(context.Background()); err != nil {
				Fail(err.Error())
			}
		})

		Context("when collection is empty", func() {
			Context("when filter is nil", func() {
				It("should return document not found error", func() {
					receivedDummy, receivedErr := sut.FirstInserted(context.Background(), nil)
					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
					Expect(receivedDummy).To(Equal(DummyStruct{}))
				})
			})

			Context("when filter is filled", func() {
				BeforeEach(func() {
					filter = map[string]any{"int": 0}
				})

				It("should return document not found error", func() {
					receivedDummy, receivedErr := sut.FirstInserted(context.Background(), filter)
					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
					Expect(receivedDummy).To(Equal(DummyStruct{}))
				})
			})
		})

		Context("when collection is filled", func() {
			var (
				dummies       []DummyStruct
				expectedDummy DummyStruct
			)

			BeforeAll(func() {
				By("populating with Create")
				documentCount := randomIntBetween(10, 20)
				dummies, err = populateCollectionWithManyFakeDocuments(sut, documentCount)
				if err != nil {
					Fail(err.Error())
				}
			})

			Context("when filter is nil", func() {
				BeforeAll(func() {
					expectedDummy = dummies[0]
				})

				It("should return first document and no error", func() {
					receivedDummy, receivedErr := sut.FirstInserted(context.Background(), nil)
					Expect(receivedErr).NotTo(HaveOccurred())
					Expect(receivedDummy).To(Equal(expectedDummy))
				})
			})

			Context("when filter is empty", func() {
				BeforeAll(func() {
					filter = map[string]any{}
					expectedDummy = dummies[0]
				})

				It("should return first document and no error", func() {
					receivedDummy, receivedErr := sut.FirstInserted(context.Background(), filter)
					Expect(receivedErr).NotTo(HaveOccurred())
					Expect(receivedDummy).To(Equal(expectedDummy))
				})
			})

			Context("when filter has non existent fields", func() {
				BeforeAll(func() {
					filter = map[string]any{"nonexistent": -1}
				})

				It("should return document not found error", func() {
					receivedDummy, receivedErr := sut.FirstInserted(context.Background(), filter)
					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
					Expect(receivedDummy).To(Equal(DummyStruct{}))
				})
			})

			Context("when filter have wrong value type", func() {
				BeforeAll(func() {
					filter = map[string]any{"string": -1}
				})

				It("should return document not found error", func() {
					receivedDummy, receivedErr := sut.FirstInserted(context.Background(), filter)
					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
					Expect(receivedDummy).To(Equal(DummyStruct{}))
				})
			})

			Context("when filter does not match any document", func() {
				BeforeAll(func() {
					filter = map[string]any{"string": ""}
				})

				It("should return document not found error", func() {
					receivedDummy, receivedErr := sut.FirstInserted(context.Background(), filter)
					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
					Expect(receivedDummy).To(Equal(DummyStruct{}))
				})
			})

			Context("when filter matches one document", func() {
				BeforeAll(func() {
					expectedDummy = dummies[len(dummies)/2]
					filter = map[string]any{"string": expectedDummy.String}
				})

				It("should return correct document and no error", func() {
					receivedDummy, receivedErr := sut.FirstInserted(context.Background(), filter)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedDummy).To(Equal(expectedDummy))
				})
			})

			Context("when filter matches multiple documents", func() {
				BeforeAll(func() {
					By("ensuring with UpdateID that there are multiple documents with the same field")
					expectedDummy = dummies[len(dummies)/2]
					notExpectedDummy := dummies[len(dummies)/2+1]

					notExpectedDummy.String = expectedDummy.String
					err := sut.UpdateID(context.Background(), notExpectedDummy.ID, notExpectedDummy)
					if err != nil {
						Fail(err.Error())
					}

					filter = map[string]any{"string": expectedDummy.String}
				})

				It("should return first document and no error", func() {
					receivedDummy, receivedErr := sut.FirstInserted(context.Background(), filter)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedDummy).To(Equal(expectedDummy))
				})
			})
		})
	})

	Describe("Last", func() {
		AfterAll(func() {
			if err := sut.Drop(context.Background()); err != nil {
				Fail(err.Error())
			}
		})

		Context("when collection is empty", func() {
			It("should return document not found error", func() {
				receivedDummy, receivedErr := sut.Last(context.Background())
				Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
				Expect(receivedDummy).To(Equal(DummyStruct{}))
			})
		})

		Context("when collection is filled", func() {
			var expectedDummy DummyStruct

			BeforeAll(func() {
				By("populating with Create")
				documentCount := randomIntBetween(10, 20)
				dummies, err := populateCollectionWithManyFakeDocuments(sut, documentCount)
				if err != nil {
					Fail(err.Error())
				}

				expectedDummy = dummies[len(dummies)-1]
			})

			It("should return last document and no error", func() {
				receivedDummy, receivedErr := sut.Last(context.Background())
				Expect(receivedErr).ToNot(HaveOccurred())
				Expect(receivedDummy).To(Equal(expectedDummy))
			})
		})
	})

	Describe("LastInserted", func() {
		var filter any

		AfterAll(func() {
			if err := sut.Drop(context.Background()); err != nil {
				Fail(err.Error())
			}
		})

		Context("when collection is empty", func() {
			Context("when filter is nil", func() {
				It("should return document not found error", func() {
					receivedDummy, receivedErr := sut.LastInserted(context.Background(), nil)
					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
					Expect(receivedDummy).To(Equal(DummyStruct{}))
				})
			})

			Context("when filter is filled", func() {
				BeforeEach(func() {
					filter = map[string]any{"int": 0}
				})

				It("should return document not found error", func() {
					receivedDummy, receivedErr := sut.LastInserted(context.Background(), filter)
					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
					Expect(receivedDummy).To(Equal(DummyStruct{}))
				})
			})
		})

		Context("when collection is filled", func() {
			var (
				dummies       []DummyStruct
				expectedDummy DummyStruct
			)

			BeforeAll(func() {
				By("populating with Create")
				documentCount := randomIntBetween(10, 20)
				dummies, err = populateCollectionWithManyFakeDocuments(sut, documentCount)
				if err != nil {
					Fail(err.Error())
				}
			})

			Context("when filter is nil", func() {
				BeforeAll(func() {
					expectedDummy = dummies[len(dummies)-1]
				})

				It("should return first document and no error", func() {
					receivedDummy, receivedErr := sut.LastInserted(context.Background(), nil)
					Expect(receivedErr).NotTo(HaveOccurred())
					Expect(receivedDummy).To(Equal(expectedDummy))
				})
			})

			Context("when filter is empty", func() {
				BeforeEach(func() {
					filter = map[string]any{}
					expectedDummy = dummies[len(dummies)-1]
				})

				It("should return first document and no error", func() {
					receivedDummy, receivedErr := sut.LastInserted(context.Background(), filter)
					Expect(receivedErr).NotTo(HaveOccurred())
					Expect(receivedDummy).To(Equal(expectedDummy))
				})
			})

			Context("when filter has non existent fields", func() {
				BeforeEach(func() {
					filter = map[string]any{"nonexistent": 0}
				})

				It("should return document not found error", func() {
					receivedDummy, receivedErr := sut.LastInserted(context.Background(), filter)
					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
					Expect(receivedDummy).To(Equal(DummyStruct{}))
				})
			})

			Context("when filter have wrong value type", func() {
				BeforeEach(func() {
					filter = map[string]any{"string": -1}
				})

				It("should return document not found error", func() {
					receivedDummy, receivedErr := sut.LastInserted(context.Background(), filter)
					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
					Expect(receivedDummy).To(Equal(DummyStruct{}))
				})
			})

			Context("when filter does not match any document", func() {
				BeforeEach(func() {
					filter = map[string]any{"string": ""}
				})

				It("should return document not found error", func() {
					receivedDummy, receivedErr := sut.LastInserted(context.Background(), filter)
					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
					Expect(receivedDummy).To(Equal(DummyStruct{}))
				})
			})

			Context("when filter matches one document", func() {
				BeforeEach(func() {
					expectedDummy = dummies[len(dummies)/2]
					filter = map[string]any{"string": expectedDummy.String}
				})

				It("should return correct document and no error", func() {
					receivedDummy, receivedErr := sut.LastInserted(context.Background(), filter)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedDummy).To(Equal(expectedDummy))
				})
			})

			Context("when filter matches multiple documents", func() {
				BeforeEach(func() {
					By("ensuring with UpdateID that there are multiple documents with the same field")
					expectedDummy = dummies[len(dummies)/2]
					notExpectedDummy := dummies[len(dummies)/2-1]

					notExpectedDummy.String = expectedDummy.String
					err := sut.UpdateID(context.Background(), notExpectedDummy.ID, notExpectedDummy)
					if err != nil {
						Fail(err.Error())
					}

					filter = map[string]any{"string": expectedDummy.String}
				})

				It("should return first document and no error", func() {
					receivedDummy, receivedErr := sut.LastInserted(context.Background(), filter)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedDummy).To(Equal(expectedDummy))
				})
			})
		})
	})

	Describe("UpdateID", func() {
		var dummy DummyStruct

		AfterAll(func() {
			if err := sut.Drop(context.Background()); err != nil {
				Fail(err.Error())
			}
		})

		Context("when id is nil", func() {
			It("should return empty id error", func() {
				receivedErr := sut.UpdateID(context.Background(), nil, DummyStruct{})
				Expect(receivedErr).To(MatchError(ErrEmptyID))
			})
		})

		Context("when collection is empty", func() {
			BeforeAll(func() {
				dummy.ID, err = notExistentID()
				if err != nil {
					Fail(err.Error())
				}
			})

			It("should return error", func() {
				receivedErr := sut.UpdateID(context.Background(), dummy.ID, DummyStruct{})
				Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
			})
		})

		Context("when collection is filled", func() {
			var dummies []DummyStruct

			BeforeAll(func() {
				By("populating with Create")
				documentCount := randomIntBetween(10, 20)
				dummies, err = populateCollectionWithManyFakeDocuments(sut, documentCount)
				if err != nil {
					Fail(err.Error())
				}
			})

			Context("when ID does not exist", func() {
				BeforeAll(func() {
					dummy.ID, err = notExistentID()
					if err != nil {
						Fail(err.Error())
					}
				})

				It("should return error and not update any document", func() {
					receivedErr := sut.UpdateID(context.Background(), dummy.ID, DummyStruct{})
					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))

					By("validating with All")
					receivedAll, receivedErr := sut.All(context.Background())
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedAll).To(Equal(dummies))
				})
			})

			Context("when ID is from first document", func() {
				BeforeAll(func() {
					if err := fakeData(&dummy); err != nil {
						Fail(err.Error())
					}

					dummy.ID = dummies[0].ID
					dummies[0] = dummy
				})

				It("should return no error and update document", func() {
					receivedErr := sut.UpdateID(context.Background(), dummy.ID, dummy)
					Expect(receivedErr).ToNot(HaveOccurred())

					By("validating with All")
					receivedAll, receivedErr := sut.All(context.Background())
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedAll).To(Equal(dummies))
				})
			})

			Context("when ID is from last document", func() {
				BeforeAll(func() {
					if err := fakeData(&dummy); err != nil {
						Fail(err.Error())
					}

					dummy.ID = dummies[len(dummies)-1].ID
					dummies[len(dummies)-1] = dummy
				})

				It("should return no error and update document", func() {
					receivedErr := sut.UpdateID(context.Background(), dummy.ID, dummy)
					Expect(receivedErr).ToNot(HaveOccurred())

					By("validating with All")
					receivedAll, receivedErr := sut.All(context.Background())
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedAll).To(Equal(dummies))
				})
			})

			Context("when ID is from a document in the middle of the collection", func() {
				BeforeAll(func() {
					if err := fakeData(&dummy); err != nil {
						Fail(err.Error())
					}

					dummy.ID = dummies[len(dummies)/2].ID
					dummies[len(dummies)/2] = dummy
				})

				It("should return no error and update document", func() {
					receivedErr := sut.UpdateID(context.Background(), dummy.ID, dummy)
					Expect(receivedErr).ToNot(HaveOccurred())

					By("validating with All")
					receivedAll, receivedErr := sut.All(context.Background())
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedAll).To(Equal(dummies))
				})
			})
		})
	})

	Describe("Where", func() {
		var filter map[string]any

		AfterAll(func() {
			if err := sut.Drop(context.Background()); err != nil {
				Fail(err.Error())
			}
		})

		Context("when collection is empty", func() {
			Context("when filter is nil", func() {
				It("should return empty slice and no error", func() {
					receivedStructs, receivedErr := sut.Where(context.Background(), nil)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(BeEmpty())
				})
			})

			Context("when filter is empty", func() {
				BeforeEach(func() {
					filter = map[string]any{}
				})

				It("should return empty slice and no error", func() {
					receivedStructs, receivedErr := sut.Where(context.Background(), filter)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(BeEmpty())
				})
			})

			Context("when filter is filled", func() {
				BeforeEach(func() {
					filter = map[string]any{"string": ""}
				})

				It("should return empty slice and no error", func() {
					receivedStructs, receivedErr := sut.Where(context.Background(), filter)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(BeEmpty())
				})
			})
		})

		Context("when collection is filled", func() {
			var (
				expectedDummies []DummyStruct
				dummies         []DummyStruct
				documentCount   int
			)

			BeforeAll(func() {
				By("populating with Create")
				documentCount = randomIntBetween(10, 20)
				dummies, err = populateCollectionWithManyFakeDocuments(sut, documentCount)
				if err != nil {
					Fail(err.Error())
				}
			})

			Context("when filter is nil", func() {
				BeforeEach(func() {
					expectedDummies = dummies
				})

				It("should return all documents and no error", func() {
					receivedStructs, receivedErr := sut.Where(context.Background(), nil)
					Expect(receivedErr).NotTo(HaveOccurred())
					Expect(receivedStructs).To(Equal(expectedDummies))
				})
			})

			Context("when filter is empty", func() {
				BeforeEach(func() {
					filter = map[string]any{}
				})

				It("should return all documents and no error", func() {
					receivedStructs, receivedErr := sut.Where(context.Background(), filter)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(Equal(dummies))
				})
			})

			Context("when filter does not have existing fields", func() {
				BeforeEach(func() {
					filter = map[string]any{"nonexistent": 0}
				})

				It("should return empty slice and no error", func() {
					receivedStructs, receivedErr := sut.Where(context.Background(), filter)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(BeEmpty())
				})
			})

			Context("when filter have wrong value type", func() {
				BeforeEach(func() {
					filter = map[string]any{"string": -1}
				})

				It("should return empty slice and no error", func() {
					receivedStructs, receivedErr := sut.Where(context.Background(), filter)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(BeEmpty())
				})
			})

			Context("when filter does not match any document", func() {
				BeforeEach(func() {
					filter = map[string]any{"string": ""}
				})

				It("should return empty slice and no error", func() {
					receivedStructs, receivedErr := sut.Where(context.Background(), filter)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(BeEmpty())
				})
			})

			Context("when filter matches one document", func() {
				BeforeEach(func() {
					expectedDummy := dummies[documentCount/2]
					expectedDummies = []DummyStruct{expectedDummy}
					filter = map[string]any{"string": expectedDummy.String}
				})

				It("should return correct documents and no error", func() {
					receivedStructs, receivedErr := sut.Where(context.Background(), filter)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(Equal(expectedDummies))
				})
			})

			Context("when filter matches multiple documents", func() {
				BeforeEach(func() {
					By("ensuring with UpdateID that there are multiple documents with the same field")
					firstExpectedDummy := dummies[documentCount/2]
					secondExpectedDummy := dummies[documentCount/2+1]

					secondExpectedDummy.String = firstExpectedDummy.String
					err := sut.UpdateID(context.Background(), secondExpectedDummy.ID, secondExpectedDummy)
					if err != nil {
						Fail(err.Error())
					}

					expectedDummies = []DummyStruct{firstExpectedDummy, secondExpectedDummy}
					filter = map[string]any{"string": firstExpectedDummy.String}
				})

				It("should return all matching documents and no error", func() {
					receivedStructs, receivedErr := sut.Where(context.Background(), filter)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(Equal(expectedDummies))
				})
			})
		})
	})

	Describe("WhereWithOrder", func() {
		var (
			filter  map[string]any
			orderBy map[string]OrderBy
		)

		AfterAll(func() {
			if err := sut.Drop(context.Background()); err != nil {
				Fail(err.Error())
			}
		})

		Context("when collection is empty", func() {
			Context("when filter is nil and order is nil", func() {
				It("should return empty slice and no error", func() {
					receivedStructs, receivedErr := sut.WhereWithOrder(context.Background(), nil, nil)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(BeEmpty())
				})
			})

			Context("when filter is nil and order is filled", func() {
				BeforeEach(func() {
					orderBy = map[string]OrderBy{"string": OrderAsc}
				})

				It("should return empty slice and no error", func() {
					receivedStructs, receivedErr := sut.WhereWithOrder(context.Background(), nil, orderBy)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(BeEmpty())
				})
			})

			Context("when filter is filled and order is nil", func() {
				BeforeEach(func() {
					filter = map[string]any{"string": ""}
				})

				It("should return empty slice and no error", func() {
					receivedStructs, receivedErr := sut.WhereWithOrder(context.Background(), filter, nil)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(BeEmpty())
				})
			})

			Context("when filter is filled and order is filled", func() {
				BeforeEach(func() {
					filter = map[string]any{"string": ""}
					orderBy = map[string]OrderBy{"string": OrderAsc}
				})

				It("should return empty slice and no error", func() {
					receivedStructs, receivedErr := sut.WhereWithOrder(context.Background(), filter, orderBy)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(BeEmpty())
				})
			})
		})

		Context("when collection is filled", func() {
			var dummies []DummyStruct

			BeforeAll(func() {
				By("populating with Create")
				documentCount := randomIntBetween(10, 20)
				dummies, err = populateCollectionWithManyFakeDocuments(sut, documentCount)
				if err != nil {
					Fail(err.Error())
				}
			})

			Context("when filter is nil and order is nil", func() {
				It("should return all documents with no order and no error", func() {
					receivedStructs, receivedErr := sut.WhereWithOrder(context.Background(), nil, nil)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(Equal(dummies))
				})
			})

			Context("when filter is nil and order is filled", func() {
				BeforeEach(func() {
					orderBy = map[string]OrderBy{"int": OrderAsc}
				})

				It("should return all documents ordered and no error", func() {
					receivedStructs, receivedErr := sut.WhereWithOrder(context.Background(), nil, orderBy)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(ContainElements(dummies))
					Expect(receivedStructs).To(HaveLen(len(dummies)))

					for i := 1; i < len(receivedStructs); i++ {
						Expect(receivedStructs[i].Int).To(BeNumerically(">=", receivedStructs[i-1].Int))
					}
				})
			})

			Context("when filter is not nil and order is nil", func() {
				BeforeEach(func() {
					filter = map[string]any{}
				})

				It("should return all documents with no order and no error", func() {
					receivedStructs, receivedErr := sut.WhereWithOrder(context.Background(), filter, nil)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(Equal(dummies))
				})
			})

			Context("when filter is not nil and order is filled", func() {
				BeforeEach(func() {
					filter = map[string]any{}
					orderBy = map[string]OrderBy{"int": OrderDesc}
				})

				It("should return all documents ordered and no error", func() {
					receivedStructs, receivedErr := sut.WhereWithOrder(context.Background(), filter, orderBy)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(ContainElements(dummies))

					for i := 1; i < len(receivedStructs); i++ {
						Expect(receivedStructs[i].Int).To(BeNumerically("<=", receivedStructs[i-1].Int))
					}
				})
			})

			Context("when filter does not have existing fields", func() {
				BeforeEach(func() {
					filter = map[string]any{"nonexistent": 0}
					orderBy = map[string]OrderBy{"int": OrderAsc}
				})

				It("should return empty slice and no error", func() {
					receivedStructs, receivedErr := sut.WhereWithOrder(context.Background(), filter, orderBy)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(BeEmpty())
				})
			})

			Context("when order does not have existing fields", func() {
				BeforeEach(func() {
					filter = map[string]any{}
					orderBy = map[string]OrderBy{"nonexistent": OrderAsc}
				})

				It("should return all documents and no error", func() {
					receivedStructs, receivedErr := sut.WhereWithOrder(context.Background(), filter, orderBy)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(Equal(dummies))
				})
			})

			Context("when filter have wrong value type", func() {
				BeforeEach(func() {
					filter = map[string]any{"string": -1}
					orderBy = map[string]OrderBy{"string": OrderAsc}
				})

				It("should return empty slice and no error", func() {
					receivedStructs, receivedErr := sut.WhereWithOrder(context.Background(), filter, orderBy)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(BeEmpty())
				})
			})

			Context("when orderby have wrong value type", func() {
				BeforeEach(func() {
					filter = map[string]any{}
					orderBy = map[string]OrderBy{"string": 0}
				})

				It("should return empty slice and no error", func() {
					receivedStructs, receivedErr := sut.WhereWithOrder(context.Background(), filter, orderBy)
					Expect(receivedErr).To(MatchError(ErrInvalidOrder))
					Expect(receivedStructs).To(BeEmpty())
				})
			})

			Context("when filter does not match any document", func() {
				BeforeEach(func() {
					filter = map[string]any{"string": ""}
					orderBy = map[string]OrderBy{"string": OrderAsc}
				})

				It("should return empty slice and no error", func() {
					receivedStructs, receivedErr := sut.WhereWithOrder(context.Background(), filter, orderBy)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(BeEmpty())
				})
			})

			Context("when filter matches one document", func() {
				var expectedDummies []DummyStruct

				BeforeEach(func() {
					expectedDummy := dummies[len(dummies)/2]
					expectedDummies = []DummyStruct{expectedDummy}

					filter = map[string]any{"string": expectedDummy.String}
					orderBy = map[string]OrderBy{"string": OrderAsc}
				})

				It("should return correct documents and no error", func() {
					receivedStructs, receivedErr := sut.WhereWithOrder(context.Background(), filter, orderBy)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(Equal(expectedDummies))
				})
			})

			Context("when filter matches multiple documents", func() {
				var expectedDummies []DummyStruct

				BeforeEach(func() {
					By("ensuring with UpdateID that there are multiple documents with the same field")
					firstExpectedDummy := dummies[len(dummies)/2]
					secondExpectedDummy := dummies[len(dummies)/2+1]

					secondExpectedDummy.String = firstExpectedDummy.String
					err := sut.UpdateID(context.Background(), secondExpectedDummy.ID, secondExpectedDummy)
					if err != nil {
						Fail(err.Error())
					}

					expectedDummies = []DummyStruct{firstExpectedDummy, secondExpectedDummy}

					filter = map[string]any{"string": firstExpectedDummy.String}
					orderBy = map[string]OrderBy{"int": OrderAsc}
				})

				It("should return all matching ordered documents and no error", func() {
					receivedStructs, receivedErr := sut.WhereWithOrder(context.Background(), filter, orderBy)
					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(ContainElements(expectedDummies))
					Expect(receivedStructs).To(HaveLen(len(expectedDummies)))

					for i := 1; i < len(receivedStructs); i++ {
						Expect(receivedStructs[i].Int).To(BeNumerically(">=", receivedStructs[i-1].Int))
					}
				})
			})
		})
	})

	Describe("ListIndexes", func() {
		var (
			defaultIndex = Index{Name: "_id_", Keys: map[string]OrderBy{"_id": OrderAsc}}
			customIndex  = Index{Name: "custom_index", Keys: map[string]OrderBy{"string": OrderAsc}}
		)

		Context("when collection has no custom index", func() {
			It("should return no indexes", func() {
				receivedIndexes, receivedErr := sut.ListIndexes(context.Background())
				Expect(receivedErr).ToNot(HaveOccurred())
				Expect(receivedIndexes).To(HaveLen(0))
			})
		})

		Context("when collection has default id index", func() {
			BeforeAll(func() {
				By("populating with Create")
				documentCount := randomIntBetween(10, 20)
				_, err := populateCollectionWithManyFakeDocuments(sut, documentCount)
				if err != nil {
					Fail(err.Error())
				}
			})

			AfterAll(func() {
				if err := sut.Drop(context.Background()); err != nil {
					Fail(err.Error())
				}
			})

			It("should return default index", func() {
				receivedIndexes, receivedErr := sut.ListIndexes(context.Background())
				Expect(receivedErr).ToNot(HaveOccurred())
				Expect(receivedIndexes).To(Equal([]Index{defaultIndex}))
			})
		})

		Context("when collection has one custom index", func() {
			BeforeAll(func() {
				if err := sut.CreateUniqueIndex(context.Background(), customIndex); err != nil {
					Fail(err.Error())
				}
			})

			AfterAll(func() {
				if err := sut.Drop(context.Background()); err != nil {
					Fail(err.Error())
				}
			})

			It("should return default index and custom index", func() {
				receivedIndexes, receivedErr := sut.ListIndexes(context.Background())
				Expect(receivedErr).ToNot(HaveOccurred())
				Expect(receivedIndexes).To(Equal([]Index{defaultIndex, customIndex}))
			})
		})
	})

	Describe("CreateUniqueIndex", func() {
		var index Index

		Context("when keys is nil", func() {
			BeforeAll(func() {
				index = Index{Name: "", Keys: nil}
			})

			It("should return error and not create index", func() {
				receivedErr := sut.CreateUniqueIndex(context.Background(), index)
				Expect(receivedErr).To(MatchError(ErrInvalidIndex))

				By("validating with ListIndexes")
				receivedIndexes, receivedErr := sut.ListIndexes(context.Background())
				Expect(receivedErr).ToNot(HaveOccurred())
				Expect(receivedIndexes).To(HaveLen(0))
			})
		})

		Context("when keys is empty", func() {
			BeforeAll(func() {
				index = Index{Name: "", Keys: map[string]OrderBy{}}
			})

			It("should return error and not create index", func() {
				receivedErr := sut.CreateUniqueIndex(context.Background(), index)
				Expect(receivedErr).To(MatchError(ErrInvalidIndex))

				By("validating with ListIndexes")
				receivedIndexes, receivedRrr := sut.ListIndexes(context.Background())
				Expect(receivedRrr).ToNot(HaveOccurred())
				Expect(receivedIndexes).To(HaveLen(0))
			})
		})

		Context("when one key is empty", func() {
			BeforeAll(func() {
				index = Index{Name: "", Keys: map[string]OrderBy{"": OrderAsc}}
			})

			It("should return error and not create index", func() {
				receivedErr := sut.CreateUniqueIndex(context.Background(), index)
				Expect(receivedErr).To(MatchError(ErrInvalidIndex))

				By("validating with ListIndexes")
				receivedIndexes, receivedErr := sut.ListIndexes(context.Background())
				Expect(receivedErr).ToNot(HaveOccurred())
				Expect(receivedIndexes).To(HaveLen(0))
			})
		})

		Context("when key order is wrong", func() {
			BeforeAll(func() {
				index = Index{Name: "", Keys: map[string]OrderBy{"string": 0}}
			})

			It("should return error and not create index", func() {
				receivedErr := sut.CreateUniqueIndex(context.Background(), index)
				Expect(receivedErr).To(MatchError(ErrInvalidIndex))

				By("validating with ListIndexes")
				receivedIndexes, receivedErr := sut.ListIndexes(context.Background())
				Expect(receivedErr).ToNot(HaveOccurred())
				Expect(receivedIndexes).To(HaveLen(0))
			})
		})

		Context("when name is empty", func() {
			BeforeAll(func() {
				index = Index{Name: "", Keys: map[string]OrderBy{"string": OrderAsc}}
			})

			AfterAll(func() {
				if err := sut.Drop(context.Background()); err != nil {
					Fail(err.Error())
				}
			})

			It("should return no error and not create index", func() {
				receivedErr := sut.CreateUniqueIndex(context.Background(), index)
				Expect(receivedErr).ToNot(HaveOccurred())

				By("validating with ListIndexes")
				receivedIndexes, receivedErr := sut.ListIndexes(context.Background())
				Expect(receivedErr).ToNot(HaveOccurred())
				Expect(receivedIndexes).To(ContainElement(Index{Name: "string_1", Keys: index.Keys}))
			})
		})

		Context("when name is filled", func() {
			BeforeAll(func() {
				index = Index{Name: "unique_string", Keys: map[string]OrderBy{"string": OrderAsc}}
			})

			AfterAll(func() {
				if err := sut.Drop(context.Background()); err != nil {
					Fail(err.Error())
				}
			})

			It("should return no error create index with custom name", func() {
				receivedErr := sut.CreateUniqueIndex(context.Background(), index)
				Expect(receivedErr).ToNot(HaveOccurred())

				By("validating with ListIndexes")
				receivedIndexes, receivedErr := sut.ListIndexes(context.Background())
				Expect(receivedErr).ToNot(HaveOccurred())
				Expect(receivedIndexes).To(ContainElement(index))
			})
		})
	})

	Describe("DeleteIndex", func() {
		var (
			defaultIndex = Index{Name: "_id_", Keys: map[string]OrderBy{"_id": OrderAsc}}
			customIndex  = Index{Name: "custom_index", Keys: map[string]OrderBy{"string": OrderAsc}}
		)

		BeforeAll(func() {
			if err := sut.CreateUniqueIndex(context.Background(), customIndex); err != nil {
				Fail(err.Error())
			}
		})

		AfterAll(func() {
			if err := sut.Drop(context.Background()); err != nil {
				Fail(err.Error())
			}
		})

		Context("when index does not exist", func() {
			It("should return error", func() {
				receivedErr := sut.DeleteIndex(context.Background(), "nonexistent")
				Expect(receivedErr).To(MatchError(ErrIndexNotFound))
			})
		})

		Context("when index is default", func() {
			It("should return error and not delete index", func() {
				receivedErr := sut.DeleteIndex(context.Background(), defaultIndex.Name)
				Expect(receivedErr).To(MatchError(ErrInvalidCommandOptions))

				By("validating with ListIndexes")
				receivedIndexes, receivedErr := sut.ListIndexes(context.Background())
				Expect(receivedErr).ToNot(HaveOccurred())
				Expect(receivedIndexes).To(ContainElement(defaultIndex))
			})
		})

		Context("when index is custom", func() {
			It("should return no error and delete index", func() {
				receivedErr := sut.DeleteIndex(context.Background(), customIndex.Name)
				Expect(receivedErr).ToNot(HaveOccurred())

				By("validating with ListIndexes")
				receivedIndexes, receivedErr := sut.ListIndexes(context.Background())
				Expect(receivedErr).ToNot(HaveOccurred())
				Expect(receivedIndexes).ToNot(ContainElement(customIndex))
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
	dummies, err := generateDummyStructs(n)
	if err != nil {
		return nil, err
	}

	if err := insertManyInCollection(collection, dummies); err != nil {
		return nil, err
	}

	return dummies, nil
}

func generateDummyStructs(n int) ([]DummyStruct, error) {
	dummies := make([]DummyStruct, n)
	for i := range dummies {
		if err := fakeData(&dummies[i]); err != nil {
			return nil, err
		}
	}

	return dummies, nil
}

func insertManyInCollection(collection collection[DummyStruct], dummies []DummyStruct) error {
	for i, dummy := range dummies {
		var err error
		dummies[i].ID, err = collection.Create(context.Background(), dummy)
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
