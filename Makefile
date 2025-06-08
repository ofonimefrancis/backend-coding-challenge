GO ?= go

export GO111MODULE = on

lint-md:
	markdownlint-cli2 '**/*.md'
.PHONY: lint-md




## Run unit tests
test-unit:
	@echo ">> running unit tests"
	@$(GO) test -gcflags=-l -coverprofile=unit.coverprofile -covermode=atomic -race ./...
