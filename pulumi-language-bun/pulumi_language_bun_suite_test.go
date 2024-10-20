package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPulumiLanguageBun(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PulumiLanguageBun Suite")
}
