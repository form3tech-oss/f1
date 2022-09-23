GO_FILES?=$$(find ./ -name '*.go' | grep -v /vendor | grep -v /template/ | grep -v /build/ | grep -v swagger-client)

test: goimportscheck
	go test ./... -v -race -failfast -p 1

require-travis-env:
ifndef TRAVIS
	$(error TRAVIS is undefined)
endif

install-goimports:
	go get golang.org/x/tools/cmd/goimports

goimports:
	@goimports -w $(GO_FILES)

goimportscheck:
	@sh -c "'$(CURDIR)/scripts/goimportscheck.sh'"
