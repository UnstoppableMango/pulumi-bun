_ := $(shell mkdir -p bin)
WORKING_DIR := $(shell git rev-parse --show-toplevel)
PROJECT     := github.com/unstoppablemango/pulumi-bun

GO_SRC := $(shell find . -type f -name '*.go')

LOCALBIN := ${WORKING_DIR}/bin
PULUMI   := ${LOCALBIN}/pulumi
GINKGO   := ${LOCALBIN}/ginkgo

build: bin/pulumi-language-bun

test: | $(GINKGO)
	$(GINKGO) run ./pulumi-language-bun

tools: $(addprefix bin/,buf pulumi)
tidy: pulumi-language-bun/go.sum

clean:
	rm -rf bin/pulumi-language-bun

bin/pulumi-language-bun: $(filter ./pulumi-language-bun/%,$(GO_SRC)) pulumi-language-bun/go.mod
	go -C pulumi-language-bun build -o ${WORKING_DIR}/$@ ./

bin/pulumi: .versions/pulumi
	curl -fsSL https://get.pulumi.com | sh -s -- --install-root ${WORKING_DIR} --version $(shell cat $<) --no-edit-path

bin/ginkgo: .versions/ginkgo
	GOBIN=${LOCALBIN} go install github.com/onsi/ginkgo/v2/ginkgo@v$(shell cat $<)

.versions/pulumi:
	curl --fail -sSL https://api.github.com/repos/pulumi/pulumi/releases/latest \
	| jq -r '.tag_name' \
	| cut -c2- \
	> $@

# I'll figure out a cleaner way later
update_%: .versions/%
	rm -f $<

$(GO_SRC:%.go=%_test.go): %_test.go: | $(GINKGO)
	cd $(dir $@) && $(GINKGO) generate $(notdir $*)

%_suite_test.go: | $(GINKGO)
	cd $(dir $*) && $(GINKGO) bootstrap

pulumi-language-bun/go.mod:
	go -C pulumi-language-bun mod init ${PROJECT}/pulumi-language-bun

pulumi-language-bun/go.sum: pulumi-language-bun/go.mod $(filter ./pulumi-language-bun/%,$(GO_SRC))
	go -C pulumi-language-bun mod tidy
