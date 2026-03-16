tidy:
	go mod tidy

lint:
	golangci-lint run --fix

fmt:
	golangci-lint fmt

test:
	go test -v -cover ./...

.PHONY: tidy lint fmt test
