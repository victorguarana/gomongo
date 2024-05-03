package gomongo

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NewWatcher", Ordered, func() {
	var (
		mongodbContainerURI string
		gomongoDatabase     Database

		databaseName   = "database_test"
		collectionName = "history_test"
	)

	BeforeAll(func() {
		_, mongodbContainerURI = runMongoContainer(context.Background())
		gomongoDatabase, _ = NewDatabase(context.Background(), ConnectionSettings{
			URI:               mongodbContainerURI,
			DatabaseName:      databaseName,
			ConnectionTimeout: time.Second,
		})
	})

	Context("when the database is empty", func() {
		It("should return an error", func() {
			receivedWatcher, receivedHistory, receivedErr := NewWatcher(Database{}, collectionName)
			Expect(receivedErr).To(MatchError(ErrConnectionNotInitialized))
			Expect(receivedWatcher).To(BeNil())
			Expect(receivedHistory).To(BeNil())
		})
	})

	Context("when the database is initialized", func() {
		It("should return watcher, history and no error", func() {
			receivedWatcher, receivedHistory, receivedErr := NewWatcher(gomongoDatabase, collectionName)
			Expect(receivedErr).NotTo(HaveOccurred())
			Expect(receivedWatcher).NotTo(BeNil())
			Expect(receivedHistory).NotTo(BeNil())
		})
	})
})

var _ = Describe("Watcher", Ordered, func() {
	var (
		sut     Watcher
		history Collection[History]

		mongodbContainerURI string
		gomongoDatabase     Database

		databaseName   = "database_test"
		collectionName = "history_test"

		watchedCollection    Collection[DummyStruct]
		notWatchedCollection Collection[DummyStruct]

		ctx, ctxCancel = context.WithCancel(context.Background())
	)

	BeforeAll(func() {
		var err error
		_, mongodbContainerURI = runMongoContainer(context.Background())
		gomongoDatabase, err = NewDatabase(context.Background(), ConnectionSettings{
			URI:               mongodbContainerURI,
			DatabaseName:      databaseName,
			ConnectionTimeout: time.Second,
		})
		if err != nil {
			Fail(err.Error())
		}

		sut, history, err = NewWatcher(gomongoDatabase, collectionName)
		if err != nil {
			Fail(err.Error())
		}

		watchedCollection, err = NewCollection[DummyStruct](gomongoDatabase, "watched_collection")
		if err != nil {
			Fail(err.Error())
		}
		notWatchedCollection, err = NewCollection[DummyStruct](gomongoDatabase, "not_watched_collection")
		if err != nil {
			Fail(err.Error())
		}
	})

	Describe("Watch", func() {
		BeforeAll(func() {
			go sut.Watch(ctx, watchedCollection.Name())
		})

		AfterAll(func() {
			ctxCancel()
		})

		Context("when not watched collection has changes", func() {
			BeforeAll(func() {
				fmt.Print("Populating not watched collection")
				_, err := populateCollectionWithManyFakeDocuments(notWatchedCollection, randomIntBetween(5, 10))
				if err != nil {
					Fail(err.Error())
				}
			})

			It("should not create history entries for not watched collection changes", func() {
				By("validating with Count")
				historyEntriesCount, err := history.Count(context.Background())
				Expect(err).NotTo(HaveOccurred())
				Expect(historyEntriesCount).To(Equal(0))
			})
		})

		Context("when watched collection has changes", func() {
			Context("when a document is inserted", func() {
				var insertedDummy DummyStruct

				BeforeAll(func() {
					dummies, err := populateCollectionWithManyFakeDocuments(watchedCollection, 1)
					if err != nil {
						Fail(err.Error())
					}
					insertedDummy = dummies[0]
				})

				It("should create history entries for watched collection changes", func() {
					historyEntries, err := history.All(context.Background())
					Expect(err).NotTo(HaveOccurred())
					Expect(historyEntries).To(HaveLen(1))
					expectedModified, err := dataToBSON(insertedDummy)
					Expect(err).NotTo(HaveOccurred())
					expectedHistory := History{
						CollectionName: watchedCollection.Name(),
						ObjectID:       insertedDummy.ID,
						Modified:       expectedModified,
						Action:         "insert",
					}
					Expect(historyEntries).To(Equal([]History{expectedHistory}))
				})
			})
		})
	})
})
