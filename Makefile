GOLANGCI_VERSION := 2.0.2

.PHONY: test
test:
	go test ./... -v -race -failfast -parallel 10 -count=1 -mod=readonly

.PHONY: tools/golangci-lint
tools/golangci-lint:
	@echo "==> Installing golangci-lint..."
	@./scripts/install-golangci-lint.sh $(GOLANGCI_VERSION)

.PHONY: lint
lint: tools/golangci-lint
	@echo "==> Running golangci-lint..."
	@tools/golangci-lint run --timeout 600s

.PHONY: lint-fix
lint-fix: tools/golangci-lint
	@echo "==> Running golangci-lint..."
	@tools/golangci-lint run --timeout 600s --fix

.PHONY: install-pkgsite
install-pkgsite:
	go install golang.org/x/pkgsite/cmd/pkgsite@latest

.PHONY: open-docs
open-docs:
	pkgsite -open .

.PHONY: build-bench
build-bench:
	go build -o ./bin/f1-bench ./benchcmd
