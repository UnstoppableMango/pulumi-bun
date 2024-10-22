package main

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	testingrpc "github.com/pulumi/pulumi/sdk/v3/proto/go/testing"
)

var _ = Describe("Language", func() {
	for _, test := range languageTests {
		It(fmt.Sprintf("should pass %s", test), func(ctx context.Context) {
			result, err := testEngine.RunLanguageTest(ctx, &testingrpc.RunLanguageTestRequest{
				Token: languageTestToken,
				Test:  test,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Success).To(BeTrueBecause("the language test passes"))
		})
	}
})
