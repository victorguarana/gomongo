package mongo

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Collection", func() {
	Describe("NewCollection", func() {
		It("Returns correct collection instance", func() {
			expectedCollection := &collection{name: "Testing"}
			receivedCollection := NewCollection("Testing")

			Expect(receivedCollection).To(Equal(expectedCollection))
		})
	})

	Describe("init", func() {
		// TODO
	})
	Describe("All", func() {
		// TODO
	})
	Describe("Create", func() {
		// TODO
	})
	Describe("Count", func() {
		// TODO
	})
	Describe("DeleteID", func() {
		// TODO
	})
	Describe("FindOne", func() {
		// TODO
	})
	Describe("DeleteID", func() {
		// TODO
	})
	Describe("FindOne", func() {
		// TODO
	})
	Describe("First", func() {
		// TODO
	})
	Describe("UpdateID", func() {
		// TODO
	})
	Describe("Where", func() {
		// TODO
	})
})
