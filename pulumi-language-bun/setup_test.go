package main

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

	. "github.com/onsi/gomega"
	"github.com/pulumi/pulumi/sdk/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/common/diag"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/rpcutil"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	testingrpc "github.com/pulumi/pulumi/sdk/v3/proto/go/testing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pbempty "google.golang.org/protobuf/types/known/emptypb"
)

type hostEngine struct {
	pulumirpc.UnimplementedEngineServer
	t *testing.T

	logLock         sync.Mutex
	logRepeat       int
	previousMessage string
}

func (e *hostEngine) Log(_ context.Context, req *pulumirpc.LogRequest) (*pbempty.Empty, error) {
	// https://github.com/pulumi/pulumi/blob/fd07f42c6116ea9685b24ca35ba5832994196fc5/sdk/nodejs/cmd/pulumi-language-nodejs/language_test.go#L51
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
		e.t.Logf("Last message repeated %d times", e.logRepeat)
	}
	e.logRepeat = 1
	e.previousMessage = message

	if req.StreamId != 0 {
		e.t.Logf("(%d) %s[%s]: %s", req.StreamId, sev, req.Urn, message)
	} else {
		e.t.Logf("%s[%s]: %s", sev, req.Urn, message)
	}

	return &pbempty.Empty{}, nil
}

func runEngine(t *testing.T) string {
	engine := &hostEngine{t: t}
	stop := make(chan bool)
	t.Cleanup(func() {
		close(stop)
	})

	handle, err := rpcutil.ServeWithOptions(rpcutil.ServeOptions{
		Cancel: stop,
		Init: func(srv *grpc.Server) error {
			pulumirpc.RegisterEngineServer(srv, engine)
			return nil
		},
		Options: rpcutil.OpenTracingServerInterceptorOptions(nil),
	})
	NewWithT(t).Expect(err).NotTo(HaveOccurred())

	return fmt.Sprintf("127.0.0.1:%v", handle.Port)
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

func runTestingHost(t *testing.T, ctx context.Context) (string, testingrpc.LanguageTestClient) {
	g := NewWithT(t)

	binary, err := prepareTestLanguage(ctx)
	g.Expect(err).NotTo(HaveOccurred())

	// https://github.com/pulumi/pulumi/blob/fd07f42c6116ea9685b24ca35ba5832994196fc5/sdk/nodejs/cmd/pulumi-language-nodejs/language_test.go#L115
	cmd := exec.Command(binary)
	stdout, err := cmd.StdoutPipe()
	g.Expect(err).NotTo(HaveOccurred())
	stderr, err := cmd.StderrPipe()
	g.Expect(err).NotTo(HaveOccurred())
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
			t.Logf("engine: %s", text)
		}
	}()

	err = cmd.Start()
	g.Expect(err).NotTo(HaveOccurred())

	stdoutBytes, err := io.ReadAll(stdout)
	g.Expect(err).NotTo(HaveOccurred())

	address := string(stdoutBytes)

	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(rpcutil.OpenTracingClientInterceptor()),
		grpc.WithStreamInterceptor(rpcutil.OpenTracingStreamClientInterceptor()),
		rpcutil.GrpcChannelOptions(),
	)
	g.Expect(err).NotTo(HaveOccurred())

	client := testingrpc.NewLanguageTestClient(conn)

	t.Cleanup(func() {
		g.Expect(cmd.Process.Kill()).To(Succeed())
		wg.Wait()
		// We expect this to error because we just killed it.
		contract.IgnoreError(cmd.Wait())
	})

	engineAddress := runEngine(t)
	return engineAddress, client
}

func runLanguagePlugin(t *testing.T, ctx context.Context, address string, engine testingrpc.LanguageTestClient) string {
	g := NewWithT(t)

	cancel := make(chan bool)
	handle, err := rpcutil.ServeWithOptions(rpcutil.ServeOptions{
		Init: func(srv *grpc.Server) error {
			host := NewLanguageHost("TODO", address, "", "")
			pulumirpc.RegisterLanguageRuntimeServer(srv, host)
			return nil
		},
		Cancel: cancel,
	})
	g.Expect(err).NotTo(HaveOccurred())

	root, err := filepath.Abs(t.TempDir())
	g.Expect(err).NotTo(HaveOccurred())

	prepare, err := engine.PrepareLanguageTests(ctx, &testingrpc.PrepareLanguageTestsRequest{
		LanguagePluginName:   "bun",
		LanguagePluginTarget: fmt.Sprintf("127.0.0.1:%d", handle.Port),
		TemporaryDirectory:   root,
		SnapshotDirectory:    "./testdata",
		CoreSdkDirectory:     "../../vendor/sdk/nodejs", // How far does this get us
		CoreSdkVersion:       sdk.Version.String(),
		//https://github.com/pulumi/pulumi/blob/fd07f42c6116ea9685b24ca35ba5832994196fc5/sdk/nodejs/cmd/pulumi-language-nodejs/language_test.go#L220
		SnapshotEdits: []*testingrpc.PrepareLanguageTestsRequest_Replacement{
			{
				Path:        "package\\.json",
				Pattern:     fmt.Sprintf("pulumi-pulumi-%s\\.tgz", sdk.Version.String()),
				Replacement: "pulumi-pulumi-CORE.VERSION.tgz",
			},
			{
				Path:        "package\\.json",
				Pattern:     filepath.Join(root, "artifacts"),
				Replacement: "ROOT/artifacts",
			},
		},
	})
	g.Expect(err).NotTo(HaveOccurred())

	t.Cleanup(func() {
		close(cancel)
		g.Expect(<-handle.Done).NotTo(HaveOccurred())
	})

	return prepare.Token
}
