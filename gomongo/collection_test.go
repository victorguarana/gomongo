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
				By("populating with Create")
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
			It("should return document not found error", func() {
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
				By("populating with Create")
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

	Describe("FindOne", func() {
		var (
			filter any
		)

		AfterAll(func() {
			if err := sut.Drop(context.Background()); err != nil {
				Fail(err.Error())
			}
		})

		Context("when collection is empty", func() {
			It("should return document not found error", func() {
				receivedDummy, receivedErr := sut.FindOne(context.Background(), nil)

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
				By("populating with Create")
				documentCount = randomIntBetween(10, 20)
				dummyStructs, err = populateCollectionWithManyFakeDocuments(sut, documentCount)
				if err != nil {
					Fail(err.Error())
				}
			})

			Context("when filter is nil", func() {
				It("should return document not found error", func() {
					receivedDummy, receivedErr = sut.FindOne(context.Background(), nil)

					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
					Expect(receivedDummy).To(Equal(DummyStruct{}))
				})
			})

			Context("when filter does not match any document", func() {
				BeforeEach(func() {
					filter = map[string]any{"string": ""}
				})

				It("should return document not found error", func() {
					receivedDummy, receivedErr = sut.FindOne(context.Background(), filter)

					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
					Expect(receivedDummy).To(Equal(DummyStruct{}))
				})
			})

			Context("when filter matches a document", func() {
				BeforeEach(func() {
					expectedDummy = dummyStructs[documentCount/2]
					filter = map[string]any{"string": expectedDummy.String}
				})

				It("should return correct document and no error", func() {
					receivedDummy, receivedErr = sut.FindOne(context.Background(), filter)

					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedDummy).To(Equal(expectedDummy))
				})
			})

			Context("when filter is empty", func() {
				BeforeEach(func() {
					expectedDummy = dummyStructs[0]
					filter = map[string]any{}
				})

				It("should return first document and no error", func() {
					receivedDummy, receivedErr = sut.FindOne(context.Background(), filter)

					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedDummy).To(Equal(expectedDummy))
				})
			})

			Context("when filter does have not existent fields", func() {
				BeforeEach(func() {
					filter = map[string]any{"not_existent": 0}
				})

				It("should return document not found error", func() {
					receivedDummy, receivedErr = sut.FindOne(context.Background(), filter)

					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
					Expect(receivedDummy).To(Equal(DummyStruct{}))
				})
			})
		})
	})

	Describe("First", func() {
		AfterAll(func() {
			if err := sut.Drop(context.Background()); err != nil {
				Fail(err.Error())
			}
		})

		Context("when collection is empty", func() {
			It("should return document not found error", func() {
				receivedDummy, receivedErr := sut.First(context.Background())

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
				By("populating with Create")
				documentCount = randomIntBetween(10, 20)
				dummyStructs, err = populateCollectionWithManyFakeDocuments(sut, documentCount)
				if err != nil {
					Fail(err.Error())
				}
			})

			Context("when collection has documents", func() {
				BeforeEach(func() {
					expectedDummy = dummyStructs[0]
				})

				It("should return first document and no error", func() {
					receivedDummy, receivedErr = sut.First(context.Background())

					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedDummy).To(Equal(expectedDummy))
				})
			})
		})
	})

	Describe("FirstInserted", func() {
		var (
			filter any
		)

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

			Context("when filter is not empty", func() {
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

		Context("when collection is not empty", func() {
			var (
				dummyStructs  []DummyStruct
				documentCount int
				receivedDummy DummyStruct
				receivedErr   error
				expectedDummy DummyStruct
			)

			BeforeAll(func() {
				By("populating with Create")
				documentCount = randomIntBetween(10, 20)
				dummyStructs, err = populateCollectionWithManyFakeDocuments(sut, documentCount)
				if err != nil {
					Fail(err.Error())
				}
			})

			Context("when filter is nil", func() {
				It("should return first document and no error", func() {
					receivedDummy, receivedErr = sut.FirstInserted(context.Background(), nil)

					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
					Expect(receivedDummy).To(Equal(DummyStruct{}))
				})
			})

			Context("when filter is empty", func() {
				BeforeEach(func() {
					filter = map[string]any{}
					expectedDummy = dummyStructs[0]
				})

				It("should return first document and no error", func() {
					receivedDummy, receivedErr = sut.FirstInserted(context.Background(), filter)

					Expect(receivedErr).NotTo(HaveOccurred())
					Expect(receivedDummy).To(Equal(expectedDummy))
				})
			})

			Context("when filter does not match any document", func() {
				BeforeEach(func() {
					filter = map[string]any{"string": ""}
				})

				It("should return document not found error", func() {
					receivedDummy, receivedErr = sut.FirstInserted(context.Background(), filter)

					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
					Expect(receivedDummy).To(Equal(DummyStruct{}))
				})
			})

			Context("when filter does have not existent fields", func() {
				BeforeEach(func() {
					filter = map[string]any{"not_existent": 0}
				})

				It("should return document not found error", func() {
					receivedDummy, receivedErr = sut.FirstInserted(context.Background(), filter)

					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
					Expect(receivedDummy).To(Equal(DummyStruct{}))
				})
			})

			Context("when filter have wrong value type", func() {
				BeforeEach(func() {
					filter = map[string]any{"string": -1}
				})

				It("should return document not found error", func() {
					receivedDummy, receivedErr = sut.FirstInserted(context.Background(), filter)

					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
					Expect(receivedDummy).To(Equal(DummyStruct{}))
				})
			})

			Context("when filter matches one document", func() {
				BeforeEach(func() {
					expectedDummy = dummyStructs[documentCount/2]
					filter = map[string]any{"string": expectedDummy.String}
				})

				It("should return correct document and no error", func() {
					receivedDummy, receivedErr = sut.FirstInserted(context.Background(), filter)

					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedDummy).To(Equal(expectedDummy))
				})
			})

			Context("when filter matches multiple documents", func() {
				BeforeEach(func() {
					By("ensuring with UpdateID that there are multiple documents with the same field")
					expectedDummy = dummyStructs[documentCount/2]
					notExpectedDummy := dummyStructs[documentCount/2+1]

					notExpectedDummy.String = expectedDummy.String
					err := sut.UpdateID(context.Background(), notExpectedDummy.ID, notExpectedDummy)
					if err != nil {
						Fail(err.Error())
					}

					filter = map[string]any{"string": expectedDummy.String}
				})

				It("should return first document and no error", func() {
					receivedDummy, receivedErr = sut.FirstInserted(context.Background(), filter)

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

		Context("when collection is not empty", func() {
			var (
				dummyStructs  []DummyStruct
				documentCount int
				receivedDummy DummyStruct
				receivedErr   error
				expectedDummy DummyStruct
			)

			BeforeAll(func() {
				By("populating with Create")
				documentCount = randomIntBetween(10, 20)
				dummyStructs, err = populateCollectionWithManyFakeDocuments(sut, documentCount)
				if err != nil {
					Fail(err.Error())
				}
			})

			Context("when collection has documents", func() {
				BeforeEach(func() {
					expectedDummy = dummyStructs[len(dummyStructs)-1]
				})

				It("should return last document and no error", func() {
					receivedDummy, receivedErr = sut.Last(context.Background())

					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedDummy).To(Equal(expectedDummy))
				})
			})
		})
	})

	Describe("LastInserted", func() {
		var (
			filter any
		)

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

			Context("when filter is not empty", func() {
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

		Context("when collection is not empty", func() {
			var (
				dummyStructs  []DummyStruct
				documentCount int
				receivedDummy DummyStruct
				receivedErr   error
				expectedDummy DummyStruct
			)

			BeforeAll(func() {
				By("populating with Create")
				documentCount = randomIntBetween(10, 20)
				dummyStructs, err = populateCollectionWithManyFakeDocuments(sut, documentCount)
				if err != nil {
					Fail(err.Error())
				}
			})

			Context("when filter is nil", func() {
				It("should return first document and no error", func() {
					receivedDummy, receivedErr = sut.LastInserted(context.Background(), nil)

					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
					Expect(receivedDummy).To(Equal(DummyStruct{}))
				})
			})

			Context("when filter is empty", func() {
				BeforeEach(func() {
					filter = map[string]any{}
					expectedDummy = dummyStructs[len(dummyStructs)-1]
				})

				It("should return first document and no error", func() {
					receivedDummy, receivedErr = sut.LastInserted(context.Background(), filter)

					Expect(receivedErr).NotTo(HaveOccurred())
					Expect(receivedDummy).To(Equal(expectedDummy))
				})
			})

			Context("when filter does not match any document", func() {
				BeforeEach(func() {
					filter = map[string]any{"string": ""}
				})

				It("should return document not found error", func() {
					receivedDummy, receivedErr = sut.LastInserted(context.Background(), filter)

					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
					Expect(receivedDummy).To(Equal(DummyStruct{}))
				})
			})

			Context("when filter does have not existent fields", func() {
				BeforeEach(func() {
					filter = map[string]any{"not_existent": 0}
				})

				It("should return document not found error", func() {
					receivedDummy, receivedErr = sut.LastInserted(context.Background(), filter)

					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
					Expect(receivedDummy).To(Equal(DummyStruct{}))
				})
			})

			Context("when filter have wrong value type", func() {
				BeforeEach(func() {
					filter = map[string]any{"string": -1}
				})

				It("should return document not found error", func() {
					receivedDummy, receivedErr = sut.LastInserted(context.Background(), filter)

					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
					Expect(receivedDummy).To(Equal(DummyStruct{}))
				})
			})

			Context("when filter matches one document", func() {
				BeforeEach(func() {
					expectedDummy = dummyStructs[documentCount/2]
					filter = map[string]any{"string": expectedDummy.String}
				})

				It("should return correct document and no error", func() {
					receivedDummy, receivedErr = sut.LastInserted(context.Background(), filter)

					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedDummy).To(Equal(expectedDummy))
				})
			})

			Context("when filter matches multiple documents", func() {
				BeforeEach(func() {
					By("ensuring with UpdateID that there are multiple documents with the same field")
					expectedDummy = dummyStructs[documentCount/2]
					notExpectedDummy := dummyStructs[documentCount/2-1]

					notExpectedDummy.String = expectedDummy.String
					err := sut.UpdateID(context.Background(), notExpectedDummy.ID, notExpectedDummy)
					if err != nil {
						Fail(err.Error())
					}

					filter = map[string]any{"string": expectedDummy.String}
				})

				It("should return first document and no error", func() {
					receivedDummy, receivedErr = sut.LastInserted(context.Background(), filter)

					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedDummy).To(Equal(expectedDummy))
				})
			})
		})
	})

	Describe("UpdateID", func() {
		var (
			updateID ID
			update   DummyStruct
		)

		AfterAll(func() {
			if err := sut.Drop(context.Background()); err != nil {
				Fail(err.Error())
			}
		})

		Context("when id is nil", func() {
			It("should return error", func() {
				receivedErr := sut.UpdateID(context.Background(), nil, DummyStruct{})

				Expect(receivedErr).To(MatchError(ErrEmptyID))
			})
		})

		Context("when collection is empty", func() {
			BeforeAll(func() {
				updateID, err = notExistentID()
				if err != nil {
					Fail(err.Error())
				}
			})

			It("should return error", func() {
				receivedErr := sut.UpdateID(context.Background(), updateID, DummyStruct{})

				Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
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

			Context("when ID does not exist", func() {
				BeforeAll(func() {
					updateID, err = notExistentID()
					if err != nil {
						Fail(err.Error())
					}
				})

				It("should return error", func() {
					receivedErr := sut.UpdateID(context.Background(), updateID, DummyStruct{})

					Expect(receivedErr).To(MatchError(ErrDocumentNotFound))
				})

				It("should not update any document", func() {
					By("validating with All")
					receivedAll, receivedErr := sut.All(context.Background())

					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedAll).To(Equal(dummyStructs))
				})
			})

			Context("when ID is from first document", func() {
				BeforeAll(func() {
					if err := fakeData(&update); err != nil {
						Fail(err.Error())
					}
					updateID = dummyStructs[0].ID
					update.ID = updateID
					dummyStructs[0] = update
				})

				It("should return no error", func() {
					receivedErr := sut.UpdateID(context.Background(), updateID, update)

					Expect(receivedErr).ToNot(HaveOccurred())
				})

				It("should update document", func() {
					By("validating with FindID")
					receivedDummy, receivedErr := sut.FindID(context.Background(), updateID)

					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedDummy).To(Equal(update))
				})

				It("should not change other documents", func() {
					By("validating with All")
					receivedAll, receivedErr := sut.All(context.Background())

					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedAll).To(Equal(dummyStructs))
				})
			})

			Context("when ID is from last document", func() {
				BeforeAll(func() {
					if err := fakeData(&update); err != nil {
						Fail(err.Error())
					}
					updateID = dummyStructs[len(dummyStructs)-1].ID
					update.ID = updateID
					dummyStructs[len(dummyStructs)-1] = update
				})

				It("should return no error", func() {
					receivedErr := sut.UpdateID(context.Background(), updateID, update)

					Expect(receivedErr).ToNot(HaveOccurred())
				})

				It("should update document", func() {
					By("validating with FindID")
					receivedDummy, receivedErr := sut.FindID(context.Background(), updateID)

					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedDummy).To(Equal(update))
				})

				It("should not change other documents", func() {
					By("validating with All")
					receivedAll, receivedErr := sut.All(context.Background())

					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedAll).To(Equal(dummyStructs))
				})
			})

			Context("when ID is from a document in the middle of the collection", func() {
				BeforeAll(func() {
					if err := fakeData(&update); err != nil {
						Fail(err.Error())
					}
					updateID = dummyStructs[len(dummyStructs)/2].ID
					update.ID = updateID
					dummyStructs[len(dummyStructs)/2] = update
				})

				It("should return no error", func() {
					receivedErr := sut.UpdateID(context.Background(), updateID, update)

					Expect(receivedErr).ToNot(HaveOccurred())
				})

				It("should update document", func() {
					By("validating with FindID")
					receivedDummy, receivedErr := sut.FindID(context.Background(), updateID)

					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedDummy).To(Equal(update))
				})

				It("should not change other documents", func() {
					By("validating with All")
					receivedAll, receivedErr := sut.All(context.Background())

					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedAll).To(Equal(dummyStructs))
				})
			})
		})
	})

	Describe("Where", func() {
		var (
			filter map[string]any
		)

		AfterAll(func() {
			if err := sut.Drop(context.Background()); err != nil {
				Fail(err.Error())
			}
		})

		Context("when collection is empty", func() {
			PContext("when filter is nil", func() {
				It("should return empty slice and no error", func() {
					receivedStructs, receivedErr := sut.Where(context.Background(), nil)

					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(BeEmpty())
				})
			})

			Context("when filter is not nil", func() {
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

		Context("when collection is not empty", func() {
			var (
				expectedDummies []DummyStruct
				dummyStructs    []DummyStruct
				documentCount   int
			)

			BeforeAll(func() {
				By("populating with Create")
				documentCount = randomIntBetween(10, 20)
				dummyStructs, err = populateCollectionWithManyFakeDocuments(sut, documentCount)
				if err != nil {
					Fail(err.Error())
				}
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

			Context("when filter matches some documents", func() {
				var (
					expectedDummy DummyStruct
				)

				BeforeEach(func() {
					expectedDummy = dummyStructs[documentCount/2]
					filter = map[string]any{"int64": expectedDummy.Int64}
				})

				It("should return matching documents and no error", func() {
					receivedStructs, receivedErr := sut.Where(context.Background(), filter)

					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(ContainElement(expectedDummy))
				})
			})

			Context("when filter is empty", func() {
				BeforeEach(func() {
					filter = map[string]any{}
				})

				It("should return all documents and no error", func() {
					receivedStructs, receivedErr := sut.Where(context.Background(), filter)

					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(Equal(dummyStructs))
				})
			})

			PContext("when filter is nil", func() {
				BeforeEach(func() {
					expectedDummies = dummyStructs
				})

				It("should return all documents and no error", func() {
					receivedStructs, receivedErr := sut.Where(context.Background(), nil)

					Expect(receivedErr).NotTo(HaveOccurred())
					Expect(receivedStructs).To(Equal(expectedDummies))
				})
			})

			Context("when filter does not have existing fields", func() {
				BeforeEach(func() {
					filter = map[string]any{"not_existent": 0}
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

			Context("when filter matches one document", func() {
				BeforeEach(func() {
					expectedDummy := dummyStructs[documentCount/2]
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
					firstExpectedDummy := dummyStructs[documentCount/2]
					secondExpectedDummy := dummyStructs[documentCount/2+1]

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
					Expect(receivedStructs).To(HaveExactElements(expectedDummies))
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
			PContext("when filter is nil and order is nil", func() {
				It("should return empty slice and no error", func() {
					receivedStructs, receivedErr := sut.WhereWithOrder(context.Background(), nil, nil)

					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(BeEmpty())
				})
			})

			PContext("when filter is nil and order is not nil", func() {
				BeforeEach(func() {
					orderBy = map[string]OrderBy{"string": OrderAsc}
				})

				It("should return empty slice and no error", func() {
					receivedStructs, receivedErr := sut.WhereWithOrder(context.Background(), nil, orderBy)

					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(BeEmpty())
				})
			})

			Context("when filter is not nil and order is nil", func() {
				BeforeEach(func() {
					filter = map[string]any{"string": ""}
				})

				It("should return empty slice and no error", func() {
					receivedStructs, receivedErr := sut.WhereWithOrder(context.Background(), filter, nil)

					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(BeEmpty())
				})
			})

			Context("when filter is not nil and order is not nil", func() {
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

		Context("when collection is not empty", func() {
			var (
				expectedDummies []DummyStruct
				dummyStructs    []DummyStruct
				documentCount   int
			)

			BeforeAll(func() {
				By("populating with Create")
				documentCount = randomIntBetween(10, 20)
				dummyStructs, err = populateCollectionWithManyFakeDocuments(sut, documentCount)
				if err != nil {
					Fail(err.Error())
				}
			})

			PContext("when filter is nil and order is nil", func() {
				It("should return empty slice and no error", func() {
					receivedStructs, receivedErr := sut.WhereWithOrder(context.Background(), nil, nil)

					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(BeEmpty())
				})
			})

			PContext("when filter is nil and order is not nil", func() {
				BeforeEach(func() {
					orderBy = map[string]OrderBy{"string": OrderAsc}
				})

				It("should return empty slice and no error", func() {
					receivedStructs, receivedErr := sut.WhereWithOrder(context.Background(), nil, orderBy)

					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(BeEmpty())
				})
			})

			Context("when filter is not nil and order is nil", func() {
				BeforeEach(func() {
					filter = map[string]any{}

					expectedDummies = dummyStructs
				})

				It("should return empty slice and no error", func() {
					receivedStructs, receivedErr := sut.WhereWithOrder(context.Background(), filter, nil)

					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(Equal(expectedDummies))
				})
			})

			Context("when filter is not nil and order is not nil", func() {
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

			Context("when filter does not have existing fields", func() {
				BeforeEach(func() {
					filter = map[string]any{"not_existent": 0}
					orderBy = map[string]OrderBy{"string": OrderAsc}
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
					orderBy = map[string]OrderBy{"not_existent": OrderAsc}

					expectedDummies = dummyStructs
				})

				It("should return empty slice and no error", func() {
					receivedStructs, receivedErr := sut.WhereWithOrder(context.Background(), filter, orderBy)

					Expect(receivedErr).ToNot(HaveOccurred())
					Expect(receivedStructs).To(Equal(expectedDummies))
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
				BeforeEach(func() {
					expectedDummy := dummyStructs[documentCount/2]
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
				BeforeEach(func() {
					By("ensuring with UpdateID that there are multiple documents with the same field")
					firstExpectedDummy := dummyStructs[documentCount/2]
					secondExpectedDummy := dummyStructs[documentCount/2+1]

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

					lastInt := receivedStructs[0].Int
					for _, receivedDummy := range receivedStructs {
						Expect(receivedDummy.Int).To(BeNumerically(">=", lastInt))
						lastInt = receivedDummy.Int
					}
				})
			})
		})
	})

	Describe("ListIndexes", func() {
		AfterAll(func() {
			if err := sut.Drop(context.Background()); err != nil {
				Fail(err.Error())
			}
		})

		Context("when collection has no custom index", func() {
			It("should return no indexes", func() {
				receivedIndexes, receivedErr := sut.ListIndexes(context.Background())

				Expect(receivedErr).ToNot(HaveOccurred())
				Expect(receivedIndexes).To(HaveLen(0))
			})
		})

		Context("when collection has one custom index", func() {
			var (
				defaultIndex    = Index{Name: "_id_", Keys: map[string]OrderBy{"_id": OrderAsc}}
				customIndex     = Index{Name: "custom_index", Keys: map[string]OrderBy{"string": OrderAsc}}
				expectedIndexes = []Index{defaultIndex, customIndex}
			)

			BeforeAll(func() {
				if err := sut.CreateUniqueIndex(context.Background(), customIndex); err != nil {
					Fail(err.Error())
				}
			})

			It("should return default index and custom index", func() {
				receivedIndexes, receivedErr := sut.ListIndexes(context.Background())

				Expect(receivedErr).ToNot(HaveOccurred())
				Expect(receivedIndexes).To(Equal(expectedIndexes))
			})
		})
	})

	Describe("CreateUniqueIndex", func() {
		var (
			index       Index
			receivedErr error
		)

		AfterAll(func() {
			if err := sut.Drop(context.Background()); err != nil {
				Fail(err.Error())
			}
		})

		Context("when keys is nil", func() {
			BeforeAll(func() {
				index = Index{Name: "", Keys: nil}

				receivedErr = sut.CreateUniqueIndex(context.Background(), index)
			})

			It("should return error", func() {
				Expect(receivedErr).To(MatchError(ErrInvalidIndex))
			})

			It("should not create index", func() {
				By("validating with ListIndexes")
				receivedIndexes, err := sut.ListIndexes(context.Background())
				if err != nil {
					Fail(err.Error())
				}

				Expect(receivedIndexes).To(HaveLen(0))
			})
		})

		Context("when keys is empty", func() {
			BeforeAll(func() {
				index = Index{Name: "", Keys: map[string]OrderBy{}}

				receivedErr = sut.CreateUniqueIndex(context.Background(), index)
			})

			It("should return error", func() {
				Expect(receivedErr).To(MatchError(ErrInvalidIndex))
			})

			It("should not create index", func() {
				By("validating with ListIndexes")
				receivedIndexes, err := sut.ListIndexes(context.Background())
				if err != nil {
					Fail(err.Error())
				}

				Expect(receivedIndexes).To(HaveLen(0))
			})
		})

		Context("when one key is empty", func() {
			BeforeAll(func() {
				index = Index{Name: "", Keys: map[string]OrderBy{"": OrderAsc}}

				receivedErr = sut.CreateUniqueIndex(context.Background(), index)
			})

			It("should return error", func() {
				Expect(receivedErr).To(MatchError(ErrInvalidIndex))
			})

			It("should not create index", func() {
				By("validating with ListIndexes")
				receivedIndexes, err := sut.ListIndexes(context.Background())
				if err != nil {
					Fail(err.Error())
				}

				Expect(receivedIndexes).To(HaveLen(0))
			})
		})

		Context("when key order is wrong", func() {
			BeforeAll(func() {
				index = Index{Name: "", Keys: map[string]OrderBy{"string": 0}}

				receivedErr = sut.CreateUniqueIndex(context.Background(), index)
			})

			It("should return error", func() {
				Expect(receivedErr).To(MatchError(ErrInvalidIndex))
			})

			It("should not create index", func() {
				By("validating with ListIndexes")
				receivedIndexes, err := sut.ListIndexes(context.Background())
				if err != nil {
					Fail(err.Error())
				}

				Expect(receivedIndexes).To(HaveLen(0))
			})
		})

		Context("when name is empty", func() {
			BeforeAll(func() {
				index = Index{Name: "", Keys: map[string]OrderBy{"string": OrderAsc}}
				receivedErr = sut.CreateUniqueIndex(context.Background(), index)
			})

			AfterAll(func() {
				if err := sut.Drop(context.Background()); err != nil {
					Fail(err.Error())
				}
			})

			It("should return no error", func() {
				Expect(receivedErr).ToNot(HaveOccurred())
			})

			It("should create index with default name", func() {
				By("validating with ListIndexes")
				receivedIndexes, err := sut.ListIndexes(context.Background())
				if err != nil {
					Fail(err.Error())
				}

				Expect(receivedIndexes).To(ContainElement(Index{Name: "string_1", Keys: index.Keys}))
			})
		})

		Context("when name is not empty", func() {
			BeforeAll(func() {
				index = Index{Name: "unique_string", Keys: map[string]OrderBy{"string": OrderAsc}}

				receivedErr = sut.CreateUniqueIndex(context.Background(), index)
			})

			AfterAll(func() {
				if err := sut.Drop(context.Background()); err != nil {
					Fail(err.Error())
				}
			})

			It("should return no error", func() {
				Expect(receivedErr).ToNot(HaveOccurred())
			})

			It("should create index wirh custom name", func() {
				By("validating with ListIndexes")
				receivedIndexes, err := sut.ListIndexes(context.Background())
				if err != nil {
					Fail(err.Error())
				}

				Expect(receivedIndexes).To(ContainElement(index))
			})
		})
	})

	Describe("DeleteIndex", func() {
		var (
			receivedErr error

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
			BeforeAll(func() {
				receivedErr = sut.DeleteIndex(context.Background(), "not_existent")
			})

			It("should return error", func() {
				Expect(receivedErr).To(MatchError(ErrIndexNotFound))
			})
		})

		Context("when index is default", func() {
			BeforeAll(func() {
				receivedErr = sut.DeleteIndex(context.Background(), defaultIndex.Name)
			})

			It("should return error", func() {
				Expect(receivedErr).To(MatchError(ErrInvalidCommandOptions))
			})

			It("should not delete index", func() {
				By("validating with ListIndexes")
				receivedIndexes, err := sut.ListIndexes(context.Background())
				if err != nil {
					Fail(err.Error())
				}

				Expect(receivedIndexes).To(ContainElement(defaultIndex))
			})
		})

		Context("when index is custom", func() {
			BeforeAll(func() {
				receivedErr = sut.DeleteIndex(context.Background(), customIndex.Name)
			})

			It("should return no error", func() {
				Expect(receivedErr).ToNot(HaveOccurred())
			})

			It("should delete index", func() {
				By("validating with ListIndexes")
				receivedIndexes, err := sut.ListIndexes(context.Background())
				if err != nil {
					Fail(err.Error())
				}

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
