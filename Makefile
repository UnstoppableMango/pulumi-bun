_ := $(shell mkdir -p bin)
WORKING_DIR := $(shell git rev-parse --show-toplevel)
PROJECT     := github.com/unstoppablemango/pulumi-bun

GO_SRC := $(shell find . -type f -name '*.go')

LOCALBIN := ${WORKING_DIR}/bin
BUF      := ${LOCALBIN}/buf
PULUMI   := ${LOCALBIN}/pulumi

build: bin/pulumi-language-bun

tools: $(addprefix bin/,buf pulumi)
proto: pulumi-language-bun/proto
tidy: pulumi-language-bun/go.sum

clean:
	rm -rf pulumi-language-bun/proto

pulumi-language-bun/proto: | $(BUF)
	$(BUF) generate

bin/pulumi-language-bun: $(filter ./pulumi-language-bun/%,$(GO_SRC))
	go -C pulumi-language-bun build -o ${WORKING_DIR}/$@ ./

bin/buf: .versions/buf
	GOBIN=${LOCALBIN} go install github.com/bufbuild/buf/cmd/buf@v$(shell cat $<)

bin/pulumi: .versions/pulumi
	curl -fsSL https://get.pulumi.com | sh -s -- --install-root ${WORKING_DIR} --version $(shell cat $<) --no-edit-path

pulumi-language-bun/go.mod:
	go -C pulumi-language-bun mod init ${PROJECT}/pulumi-language-bun

pulumi-language-bun/go.sum: pulumi-language-bun/go.mod $(filter ./pulumi-language-bun/%,$(GO_SRC))
	go -C pulumi-language-bun mod tidy
