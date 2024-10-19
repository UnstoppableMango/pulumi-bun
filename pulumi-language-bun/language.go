package main

import (
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type bunLanguageHost struct {
	pulumirpc.UnimplementedLanguageRuntimeServer

	exec                 string
	engineAddress        string
	tracing              string
	binary               string
	dotnetBuildSucceeded bool
}

func newLanguageHost(exec, engineAddress, tracing string, binary string) pulumirpc.LanguageRuntimeServer {
	return &bunLanguageHost{
		exec:          exec,
		engineAddress: engineAddress,
		tracing:       tracing,
		binary:        binary,
	}
}
