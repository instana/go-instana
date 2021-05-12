LINTER ?= $(shell go env GOPATH)/bin/golangci-lint

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
	| sh -s -- -b $(basename $(GOPATH))/bin v1.23.8

.PHONY: test vendor
