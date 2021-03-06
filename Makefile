
.PHONY: lint
lint:
	@golangci-lint run ./...

.PHONY: test
test: lint
	@go test ./... -timeout 2s -race -cover -coverprofile=./coverage.txt -covermode=atomic

.PHONY: coverage
coverage: test
	@go tool cover -html=./coverage.txt -o coverage.html

