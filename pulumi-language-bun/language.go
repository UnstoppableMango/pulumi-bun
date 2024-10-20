package main

import (
	"context"
	"fmt"
	"io"

	"github.com/pulumi/pulumi/sdk/v3/go/common/util/logging"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/rpcutil"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/spf13/afero"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type bunLanguageHost struct {
	pulumirpc.UnimplementedLanguageRuntimeServer

	exec          string
	engineAddress string
	tracing       string
	binary        string
	fsys          afero.Fs
}

func NewLanguageHost(exec, engineAddress, tracing string, binary string) pulumirpc.LanguageRuntimeServer {
	return &bunLanguageHost{
		exec:          exec,
		engineAddress: engineAddress,
		tracing:       tracing,
		binary:        binary,
		fsys:          afero.NewOsFs(),
	}
}

func (host *bunLanguageHost) connectToEngine() (pulumirpc.EngineClient, io.Closer, error) {
	conn, err := grpc.NewClient(host.engineAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		rpcutil.GrpcChannelOptions(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("language host could not make connection to engine: %w", err)
	}

	engineClient := pulumirpc.NewEngineClient(conn)
	return engineClient, conn, nil
}

func (host *bunLanguageHost) GetRequiredPlugins(
	ctx context.Context,
	req *pulumirpc.GetRequiredPluginsRequest,
) (*pulumirpc.GetRequiredPluginsResponse, error) {
	plugins, err := getPlugins(host.fsys, req.Info.ProgramDirectory)
	if err != nil {
		logging.V(3).Infof("one or more errors while discovering plugins: %s", err)
	}

	return &pulumirpc.GetRequiredPluginsResponse{
		Plugins: plugins,
	}, nil
}
