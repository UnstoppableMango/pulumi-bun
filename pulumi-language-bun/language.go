package main

import (
	"context"
	"fmt"
	"io"

	"github.com/pulumi/pulumi/sdk/v3/go/common/util/rpcutil"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

func (host *bunLanguageHost) connectToEngine() (pulumirpc.EngineClient, io.Closer, error) {
	// Make a connection to the real engine that we will log messages to.
	conn, err := grpc.Dial(
		host.engineAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		rpcutil.GrpcChannelOptions(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("language host could not make connection to engine: %w", err)
	}

	// Make a client around that connection.
	engineClient := pulumirpc.NewEngineClient(conn)
	return engineClient, conn, nil
}

func (host *bunLanguageHost) GetRequiredPlugins(
	ctx context.Context,
	req *pulumirpc.GetRequiredPluginsRequest,
) (*pulumirpc.GetRequiredPluginsResponse, error) {
	panic("unimplemented")
}
