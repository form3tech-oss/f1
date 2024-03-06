GOLANGCI_VERSION := 1.56.2

.PHONY: test
test:
	go test ./... -v -race -failfast -p 1 -mod=readonly

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


