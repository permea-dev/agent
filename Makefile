.PHONY: build test lint run tidy test-state test-transport test-pricing test-config
build:
	go build -o bin/permea ./cmd/permea
test:
	go test ./...
test-state:
	go test ./internal/state
test-transport:
	go test ./internal/transport
test-pricing:
	go test ./internal/pricing
test-config:
	go test ./internal/config
lint:
	golangci-lint run
run:
	go run ./cmd/permea --scan internal/ingest/testdata/claude_code_sample.jsonl
tidy:
	go mod tidy
