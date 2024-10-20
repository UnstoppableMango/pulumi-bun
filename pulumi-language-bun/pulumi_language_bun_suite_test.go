package main_test

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pulumi/pulumi/sdk/v3/go/common/diag"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/rpcutil"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	testingrpc "github.com/pulumi/pulumi/sdk/v3/proto/go/testing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pbempty "google.golang.org/protobuf/types/known/emptypb"
)

var (
	engine         *hostEngine
	engineAddress  string
	stopHostEngine chan bool
	host           *testHost
)

func TestPulumiLanguageBun(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pulumi Bun Language Suite")
}

var _ = BeforeSuite(func(ctx context.Context) {
	testBinary, err := prepareTestLanguage(ctx)
	Expect(err).NotTo(HaveOccurred())

	host = &testHost{}
	err = host.run(ctx, testBinary)
	Expect(err).NotTo(HaveOccurred())

	engine = &hostEngine{}
	handle, err := rpcutil.ServeWithOptions(rpcutil.ServeOptions{
		Cancel: stopHostEngine,
		Init: func(srv *grpc.Server) error {
			pulumirpc.RegisterEngineServer(srv, engine)
			return nil
		},
		Options: rpcutil.OpenTracingServerInterceptorOptions(nil),
	})

	Expect(err).NotTo(HaveOccurred())
	engineAddress = fmt.Sprintf("127.0.0.1:%v", handle.Port)
})

var _ = AfterSuite(func() {
	if stopHostEngine != nil {
		close(stopHostEngine)
	} else {
		fmt.Fprintln(GinkgoWriter, "engine stop chan was nil")
	}

	if host != nil {
		Expect(host.stop()).To(Succeed())
	} else {
		fmt.Fprintln(GinkgoWriter, "test host was nil")
	}
})

type hostEngine struct {
	pulumirpc.UnimplementedEngineServer

	logger          io.Writer
	logLock         sync.Mutex
	logRepeat       int
	previousMessage string
}

func (e *hostEngine) Log(_ context.Context, req *pulumirpc.LogRequest) (*pbempty.Empty, error) {
	e.logLock.Lock()
	defer e.logLock.Unlock()

	var sev diag.Severity
	switch req.Severity {
	case pulumirpc.LogSeverity_DEBUG:
		sev = diag.Debug
	case pulumirpc.LogSeverity_INFO:
		sev = diag.Info
	case pulumirpc.LogSeverity_WARNING:
		sev = diag.Warning
	case pulumirpc.LogSeverity_ERROR:
		sev = diag.Error
	default:
		return nil, fmt.Errorf("Unrecognized logging severity: %v", req.Severity)
	}

	message := req.Message
	if os.Getenv("PULUMI_LANGUAGE_TEST_SHOW_FULL_OUTPUT") != "true" {
		if len(message) > 1024 {
			message = message[:1024] + "... (truncated, run with PULUMI_LANGUAGE_TEST_SHOW_FULL_OUTPUT=true to see full logs))"
		}
	}

	if e.previousMessage == message {
		e.logRepeat++
		return &pbempty.Empty{}, nil
	}

	if e.logRepeat > 1 {
		fmt.Fprintf(e.logger, "Last message repeated %d times", e.logRepeat)
	}
	e.logRepeat = 1
	e.previousMessage = message

	if req.StreamId != 0 {
		fmt.Fprintf(e.logger, "(%d) %s[%s]: %s", req.StreamId, sev, req.Urn, message)
	} else {
		fmt.Fprintf(e.logger, "%s[%s]: %s", sev, req.Urn, message)
	}

	return &pbempty.Empty{}, nil
}

func prepareTestLanguage(ctx context.Context) (string, error) {
	gitbin, err := exec.LookPath("git")
	if err != nil {
		return "", fmt.Errorf("which git: %w", err)
	}

	gobin, err := exec.LookPath("go")
	if err != nil {
		return "", fmt.Errorf("which go: %w", err)
	}

	revParse, err := exec.CommandContext(ctx,
		gitbin, "rev-parse", "--show-toplevel",
	).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("searching for repo: %w", err)
	}

	repoPath := strings.TrimSpace(string(revParse))
	srcPath := filepath.Join(repoPath, "vendor", "pulumi", "cmd", "pulumi-test-language")
	binary := filepath.Join(repoPath, "bin", "pulumi-test-language")

	goBuild, err := exec.CommandContext(ctx,
		gobin, "-C", srcPath, "build", "-o", binary,
	).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("building test language: %s", goBuild)
	}

	return binary, nil
}

type testHost struct {
	cmd    *exec.Cmd
	client testingrpc.LanguageTestClient
	wg     *sync.WaitGroup
}

func (host *testHost) run(ctx context.Context, binary string) error {
	cmd := exec.CommandContext(ctx, binary)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("connecting to stdout: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("connecting to stderr: %w", err)
	}

	stderrReader := bufio.NewReader(stderr)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for {
			text, err := stderrReader.ReadString('\n')
			if err != nil {
				wg.Done()
				return
			}

			fmt.Fprintf(GinkgoWriter, "engine: %s", text)
		}
	}()

	if err = cmd.Start(); err != nil {
		return fmt.Errorf("starting test host: %w", err)
	}

	stdoutBytes, err := io.ReadAll(stdout)
	if err != nil {
		return fmt.Errorf("reading test host address: %w", err)
	}

	address := string(stdoutBytes)
	conn, err := grpc.Dial(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(rpcutil.OpenTracingClientInterceptor()),
		grpc.WithStreamInterceptor(rpcutil.OpenTracingStreamClientInterceptor()),
		rpcutil.GrpcChannelOptions(),
	)
	if err != nil {
		return fmt.Errorf("dialing test host: %w", err)
	}

	host.cmd = cmd
	host.client = testingrpc.NewLanguageTestClient(conn)
	host.wg = &wg

	return nil
}

func (host *testHost) stop() error {
	if err := host.cmd.Process.Kill(); err != nil {
		return fmt.Errorf("stopping test host: %w", err)
	}

	host.wg.Wait()

	// This will error becuase the process has stopped
	_ = host.cmd.Wait()

	return nil
}
