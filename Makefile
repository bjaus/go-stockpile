## help| Show this help dialogue
.PHONY: help
help:
	@sed -n 's/^##//p' Makefile | column -t -c 2 -s '|'

## test| Run all test cases locally
.PHONY: test
test:
	@$(GO_ENV) go test -p=1 -race -cover ./... | grep -v "no test files"

## lint| Run linters
.PHONY: lint
lint:
	@golangci-lint run

## mod| Run tidy & vendor
.PHONY: mod
mod: tidy vendor

## tidy| Update dependencies
.PHONY: tidy
tidy:
	@$(GO_ENV) go mod tidy -v

## vendor| Vendor dependencies
.PHONY: vendor
vendor:
	@$(GO_ENV) go mod vendor

