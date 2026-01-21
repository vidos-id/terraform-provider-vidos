.PHONY: fmt vet lint test cover coverprofile coverhtml coverage-gate coverage build

GOLANGCI_LINT_VERSION ?= latest
GO_TEST_COVERAGE_VERSION ?= v2.18.3

fmt:
	gofmt -w $(shell find . -name '*.go' -not -path './vendor/*')

vet:
	go vet ./...

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found; running via 'go run'"; \
		go run github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./...; \
	fi

test:
	go test ./...

cover:
	go test ./... -cover

coverprofile:
	go test ./... -coverpkg=./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

coverage-gate:
	go run github.com/vladopajic/go-test-coverage/v2@$(GO_TEST_COVERAGE_VERSION) --config .testcoverage.yml

coverage: coverprofile coverage-gate

coverhtml: coverprofile
	go tool cover -html=coverage.out -o coverage.html

build:
	go build -o ./bin/terraform-provider-vidos .
