package main_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	testingrpc "github.com/pulumi/pulumi/sdk/v3/proto/go/testing"
)

var _ = Describe("Language", func() {
	BeforeEach(func(ctx context.Context) {
		tests, err := host.client.GetLanguageTests(ctx,
			&testingrpc.GetLanguageTestsRequest{},
		)
		Expect(err).NotTo(HaveOccurred())

	})

	Context("pulumi-test-language", func() {
		BeforeEach(func() {
			Expect(engine).NotTo(BeNil())
			Expect(engineAddress).NotTo(BeEmpty())
		})
	})
})
