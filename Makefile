.PHONY: test
test:
	@go test ./... -cover -coverprofile=./coverage.txt -covermode=atomic

.PHONY: coverage
coverage: test
	@go tool cover -html=./coverage.txt -o coverage.html

