.PHONY: build test lint run tidy
build:
	go build -o bin/agente ./cmd/agente
test:
	go test ./...
lint:
	golangci-lint run
run:
	go run ./cmd/agente --scan internal/ingest/testdata/claude_code_sample.jsonl
tidy:
	go mod tidy
