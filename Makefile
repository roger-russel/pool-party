.PHONY: test
test:
	@go test ./... -timeout 5s -race -cover -coverprofile=./coverage.txt -covermode=atomic

.PHONY: coverage
coverage: test
	@go tool cover -html=./coverage.txt -o coverage.html

