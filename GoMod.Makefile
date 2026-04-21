#!Makefile

clean-mod:
	@echo "==> Clean with flag -modcache"
	@go clean -modcache
	@echo "==> Done!"; \
		echo "Please excute 'make fetch-module' or 'go mod download' ..."

fetch-mod:
	@echo "==> Fetch Go module..."; \
		go mod tidy

init: clean-mod
	@echo "==> Initializing Go module..."
	@rm -rf go.mod go.sum vendor/
	@echo 'module github.com/golang-devkit/pkg' > go.mod
	@echo '' >> go.mod
	@echo 'go 1.25.7' >> go.mod
	@echo '' >> go.mod
	@echo '' >> go.mod
	@echo '' >> go.sum
	@echo "==> Fetch Go module..."; \
		go mod tidy
	@echo "✅ Fetch Go module completed!"

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
		govulncheck \
			-show version \
			-C $(shell pwd) ./... #Please use flag "-show verbose" to show details
	@echo "✅ Successful!"

