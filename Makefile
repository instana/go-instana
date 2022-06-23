LINTER ?= $(shell which golangci-lint)

all: test legal test-build

test-build:
	rm -rf testdata/tmp
	GO_INSTANA_TEST_DUMP=1 make test
	cd testdata && go generate ./...

ifeq ($(RUN_LINTER),yes)
test: $(LINTER)
endif

test:
	go get -d -t ./... && go test $(GOFLAGS) ./...
ifeq ($(RUN_LINTER),yes)
	$(LINTER) run
endif

$(LINTER):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/a2bc9b7a99e3280805309d71036e8c2106853250/install.sh \
	| sh -s -- -b $(basename $(GOPATH))/bin v1.46.2

# Make sure there is a copyright at the first line of each .go file
legal:
	awk 'FNR==1 { if (tolower($$0) !~ "^//.+copyright") { print FILENAME" does not contain copyright header"; rc=1 } }; END { exit rc }' $$(find . -name '*.go' -type f | grep -v "_test.go")


.PHONY: test legal
