.PHONY: help
help: ## print make targets 
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: get-install-air
get-install-air: ## Installs the air build reload system using cUrl
	curl -sSfL https://raw.githubusercontent.com/air-verse/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

.PHONY: go-install-templ
go-install-templ: ## Installs the templ Templating system for Go
	go install github.com/a-h/templ/cmd/templ@latest	

.PHONY: get-install-tailwindcss
get-install-tailwindcss: ## Installs the tailwindcss-linux-x64 cli
	curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v3.4.13/tailwindcss-linux-x64
	chmod +x tailwindcss-linux-x64
	mv tailwindcss-linux-x64 tailwindcss

.PHONY: get-install-tailwindcss-arm
get-install-tailwindcss-arm: ## Installs the tailwindcss-linux-arm64 cli
	curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v3.4.13/tailwindcss-linux-arm64
	chmod +x tailwindcss-linux-arm64
	mv tailwindcss-linux-arm64 tailwindcss
