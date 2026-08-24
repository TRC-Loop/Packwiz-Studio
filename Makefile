# Packwiz Studio
#
# Cross compiling a Fyne app needs a C toolchain for the target, so the
# release targets here build for the host only. Use fyne-cross for other
# platforms.

BINARY  := packwiz-studio
PKG     := ./cmd/packwiz-studio
DIST    := dist
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

GOFILES := $(shell find . -name '*.go' -not -path './.git/*')

.DEFAULT_GOAL := build

.PHONY: help
help: ## Show the available targets
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  %-12s %s\n", $$1, $$2}'

.PHONY: build
build: ## Build the app for this machine
	go build -o $(BINARY) $(PKG)

.PHONY: run
run: ## Build and run the app
	go run $(PKG)

.PHONY: release
release: ## Build a stripped binary into dist
	@mkdir -p $(DIST)
	go build -trimpath -ldflags "-s -w" -o $(DIST)/$(BINARY) $(PKG)

.PHONY: package
package: ## Bundle a native app package with the fyne tool
	@command -v fyne >/dev/null || { \
		echo "fyne tool not installed: go install fyne.io/tools/cmd/fyne@latest"; \
		exit 1; }
	fyne package --release --src $(PKG)

.PHONY: check
check: fmt vet ## Format and vet everything

.PHONY: fmt
fmt: ## Report files that need gofmt
	@out=$$(gofmt -l $(GOFILES)); \
	if [ -n "$$out" ]; then echo "needs gofmt:"; echo "$$out"; exit 1; fi

.PHONY: fmt-fix
fmt-fix: ## Rewrite files with gofmt
	gofmt -w $(GOFILES)

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: lint
lint: ## Check the project's own rules: file length, dashes, colour literals
	@fail=0; \
	long=$$(for f in $(GOFILES); do \
		n=$$(wc -l < $$f); \
		if [ $$n -gt 200 ]; then echo "  $$f: $$n lines"; fi; \
	done); \
	if [ -n "$$long" ]; then echo "over 200 lines:"; echo "$$long"; fail=1; fi; \
	dashes=$$(grep -rn '—\|–' $(GOFILES) || true); \
	if [ -n "$$dashes" ]; then echo "em or en dashes:"; echo "$$dashes"; fail=1; fi; \
	colours=$$(grep -rn 'color\.NRGBA{\|color\.RGBA{' \
		$$(echo $(GOFILES) | tr ' ' '\n' | grep -v 'ui/tokens/') || true); \
	if [ -n "$$colours" ]; then \
		echo "colour literals outside internal/ui/tokens:"; echo "$$colours"; fail=1; fi; \
	exit $$fail

.PHONY: tidy
tidy: ## Tidy go.mod and go.sum
	go mod tidy

.PHONY: deps
deps: ## List the direct dependencies
	@go list -f '{{range .Imports}}{{.}}{{"\n"}}{{end}}' ./... \
		| grep -v '^github.com/TRC-Loop/Packwiz-Studio' \
		| grep '\.' | sort -u

.PHONY: clean
clean: ## Remove build output
	rm -rf $(DIST) $(BINARY)

.PHONY: version
version: ## Print the version the build would stamp
	@echo $(VERSION)
