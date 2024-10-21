package main

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	testingrpc "github.com/pulumi/pulumi/sdk/v3/proto/go/testing"
)

func TestPulumiLanguageBun(t *testing.T) {
	RegisterFailHandler(Fail)

	g := NewWithT(t)
	ctx := context.Background()

	engineAddress, engine := runTestingHost(t, ctx)
	tests, err := engine.GetLanguageTests(ctx,
		&testingrpc.GetLanguageTestsRequest{},
	)
	g.Expect(err).NotTo(HaveOccurred())

	RunSpecs(t, "Pulumi Bun Language Suite")
}

var _ = BeforeSuite(func(ctx context.Context) {
})

var _ = AfterSuite(func() {
})
