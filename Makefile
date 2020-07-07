GO_FILES?=$$(find ./ -name '*.go' | grep -v /vendor | grep -v /template/ | grep -v /build/ | grep -v swagger-client)

test: goimportscheck
	GOFLAGS='-mod=vendor' go test ./... -v -race -failfast -p 1

scan-code: require-snyk-token-env require-travis-env aws-auth-eu-west-1
	@echo Scanning tag $(DOCKER_TAG)
	@exec docker run --rm -v $(shell pwd):/code -e GOFLAGS='-mod=vendor' -e TRAVIS -e REPO=form3tech-oss/f1 -e SNYK_TOKEN=${SNYK_TOKEN} "288840537196.dkr.ecr.eu-west-1.amazonaws.com/tech.form3/secscan-go-golang:latest"

aws-auth-%:
	@eval $$(aws ecr get-login --region $* --no-include-email)

require-snyk-token-env:
ifndef SNYK_TOKEN
	$(error SNYK_TOKEN is undefined)
endif

require-travis-env:
ifndef TRAVIS
	$(error TRAVIS is undefined)
endif

install-goimports: 
	GO111MODULE=off go get golang.org/x/tools/cmd/goimports

goimports:
	@goimports -w $(GO_FILES)

goimportscheck:
	@sh -c "'$(CURDIR)/scripts/goimportscheck.sh'"
