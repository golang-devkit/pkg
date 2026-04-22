#!Makefile

# ====================================================================================
# Variables
# ====================================================================================
SHELL = /usr/bin/env bash

GO_TOOLCHAIN = $(shell go version | awk '{print $$3}')

clean-mod:
	@echo "==> Clean with flag -modcache"
	@go clean -modcache
	@echo "==> Done!"; \
		echo "Please execute 'make fetch-module' or 'go mod download' ..."

fetch-mod:
	@echo "==> Fetch Go module..."; \
		go mod tidy

init: clean-mod
	@echo "==> Remove go module exist..."; \
		rm -rf go.mod go.sum vendor/
	@echo "==> Initializing Go module..."; \
		go mod init github.com/golang-devkit/pkg; \
		go mod edit -go=1.25.7; \
		go mod edit -toolchain=$(GO_TOOLCHAIN);
	# Add any replace directives here:
	# @go mod edit -replace=old/path=new/path
	@echo "==> Fetch Go module..."; \
		go mod tidy
	@echo "✅ Fetch Go module completed!"

# Add any replace Use 'go get -u' to ...  directives here:
# Use 'go get ' to ...  directives here:
upgrade-module:
	@echo "==> Upgrading required packages to latest version"; \
		go get -u ./...; \
		go mod tidy
	@echo "✅ Upgrade completed!"

upgrade-module-all:
	@echo "==> Upgrading required packages and all dependency to latest version"; \
		go get -u all; \
		go mod tidy
	@echo "✅ Upgrade completed!"

fetch-module: fetch-mod upgrade-module
	@echo "==> Create vendor directory..."; \
		go mod vendor && echo "✅ Fetch Go module completed!"
	@echo "==> Run govulncheck..."; \
		go run golang.org/x/vuln/cmd/govulncheck@latest \
			-show version \
			-C $(shell pwd) ./... #Please use flag "-show verbose" to show details
	@echo "✅ Successful!"

