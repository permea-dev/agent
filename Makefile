.PHONY: build test lint run tidy
build:
	go build -o bin/permea ./cmd/permea
test:
	go test ./...
lint:
	golangci-lint run
run:
	go run ./cmd/permea --scan internal/ingest/testdata/claude_code_sample.jsonl
tidy:
	go mod tidy
