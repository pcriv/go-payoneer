tidy:
	go mod tidy

lint:
	golangci-lint run --fix

fmt:
	golangci-lint fmt

test:
	go test -v -cover ./...

release:
	@if [ -n "$$(git status --porcelain)" ]; then echo "Error: working tree is dirty" && exit 1; fi
	$(eval VERSION ?= $(shell git tag --sort=-v:refname | head -1 | sed 's/^v//' | awk -F. '{print $$1"."$$2"."$$3+1}'))
	@echo "Tagging v$(VERSION) and pushing..."
	git tag "v$(VERSION)"
	git push origin "v$(VERSION)"

.PHONY: tidy lint fmt test release
