.PHONY: lint test

lint:
	@type golangci-lint > /dev/null || go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
	golangci-lint -E gofmt run ./...

test:
	go test ./...

coverage:
	go test -covermode=count -coverprofile=coverage.txt ./...
	@type goveralls > /dev/null || go get -u github.com/mattn/goveralls
	goveralls -coverprofile=coverage.txt -service=travis-ci
