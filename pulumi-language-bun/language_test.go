package main_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Language", func() {
	Context("pulumi-test-language", func() {
		It("should setup", func() {
			Expect(engine).NotTo(BeNil())
			Expect(engineAddress).NotTo(BeEmpty())
		})
	})
})
