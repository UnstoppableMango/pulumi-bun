package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/common/util/cmdutil"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/logging"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/rpcutil"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"google.golang.org/grpc"
)

func main() {
	var tracing string
	var binary string
	var root string
	flag.StringVar(&tracing, "tracing", "", "Emit tracing to a Zipkin-compatible tracing endpoint")
	flag.StringVar(&binary, "binary", "", "A relative or an absolute path to a precompiled Bun binary to execute")
	flag.StringVar(&root, "root", "", "Project root path to use")

	var givenExecutor string
	flag.StringVar(&givenExecutor, "use-executor", "",
		"Use the given program as the executor instead of looking for one on PATH")

	flag.Parse()
	args := flag.Args()
	logging.InitLogging(false, 0, false)
	cmdutil.InitTracing("pulumi-language-bun", "pulumi-language-bun", tracing)

	var bunExec string
	switch {
	case givenExecutor != "":
		logging.V(3).Infof("language host asked to use specific executor: `%s`", givenExecutor)
		bunExec = givenExecutor
	case binary != "":
		logging.V(3).Info("language host requires no .NET SDK for a self-contained binary")
	default:
		pathExec, err := exec.LookPath("bun")
		if err != nil {
			err = fmt.Errorf("could not find `bun` on the $PATH: %w", err)
			cmdutil.Exit(err)
		}

		logging.V(3).Infof("language host identified executor from path: `%s`", pathExec)
		bunExec = pathExec
	}

	var engineAddress string
	if len(args) > 0 {
		engineAddress = args[0]
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	cancelChannel := make(chan bool)
	go func() {
		<-ctx.Done()
		cancel() // remove the interrupt handler
		close(cancelChannel)
	}()

	err := rpcutil.Healthcheck(ctx, engineAddress, 5*time.Minute, cancel)
	if err != nil {
		err = fmt.Errorf("could not start health check host RPC server: %w", err)
		cmdutil.Exit(err)
	}

	handle, err := rpcutil.ServeWithOptions(rpcutil.ServeOptions{
		Cancel: cancelChannel,
		Init: func(srv *grpc.Server) error {
			host := NewLanguageHost(bunExec, engineAddress, tracing, binary)
			pulumirpc.RegisterLanguageRuntimeServer(srv, host)
			return nil
		},
		Options: rpcutil.OpenTracingServerInterceptorOptions(nil),
	})
	if err != nil {
		err = fmt.Errorf("could not start language host RPC server: %w", err)
		cmdutil.Exit(err)
	}

	fmt.Printf("%d\n", handle.Port)

	if err := <-handle.Done; err != nil {
		err = fmt.Errorf("language host RPC stopped serving: %w", err)
		cmdutil.Exit(err)
	}
}
