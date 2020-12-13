.PHONY: test
test:
	@go test ./... -cover -coverprofile=./coverage.txt -covermode=atomic

