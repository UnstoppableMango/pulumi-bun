package main_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pulumi/pulumi/sdk/v3/go/common/diag"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/rpcutil"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"google.golang.org/grpc"
	pbempty "google.golang.org/protobuf/types/known/emptypb"
)

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

func prepareTestLanguage(ctx context.Context) error {
	gitbin, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("which git: %w", err)
	}

	gobin, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("which go: %w", err)
	}

	revParse, err := exec.CommandContext(ctx,
		gitbin, "rev-parse", "--show-toplevel",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("searching for repo: %w", err)
	}

	repo := strings.TrimSpace(string(revParse))
	versionPath := filepath.Join(repo, ".versions", "pulumi")
	readVersion, err := os.ReadFile(versionPath)
	if err != nil {
		return fmt.Errorf("reading pulumi version: %w", err)
	}

	pVersion := strings.TrimSpace(string(readVersion))

	tmp, err := os.MkdirTemp("", "")
	if err != nil {
		return fmt.Errorf("creating workdir: %w", err)
	}

	// TODO: We can switch to a submodule later and still use most of
	// this, using git to check the version matches instead of cloning.
	gitClone, err := exec.CommandContext(ctx,
		gitbin, "clone", "https://github.com/pulumi/pulumi",
	).CombinedOutput()

	return nil
}

var _ = Describe("Language", func() {
	Context("pulumi-test-language", Ordered, func() {
		var (
			engine        *hostEngine
			engineAddress string
			stop          chan bool
		)

		BeforeAll(func() {
			engine := &hostEngine{}
			handle, err := rpcutil.ServeWithOptions(rpcutil.ServeOptions{
				Cancel: stop,
				Init: func(srv *grpc.Server) error {
					pulumirpc.RegisterEngineServer(srv, engine)
					return nil
				},
				Options: rpcutil.OpenTracingServerInterceptorOptions(nil),
			})

			Expect(err).NotTo(HaveOccurred())
			engineAddress = fmt.Sprintf("127.0.0.1:%v", handle.Port)
		})

		AfterAll(func() {
			close(stop)
		})

		It("should setup", func() {
			Expect(engine).NotTo(BeNil())
			Expect(engineAddress).NotTo(BeEmpty())
		})
	})
})
