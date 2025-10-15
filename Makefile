.PHONY: help lint tidy update vendor run build

help: ## Show this help
	@egrep -h '\s##\s' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-12s\033[0m %s\n", $$1, $$2}'

lint: tidy ## Run linters
	golines -w -m 100 --base-formatter=gofumpt .
	go vet ./...

tidy: ## Tidy up dependencies
	go mod tidy

update: ## Update dependencies
	go get -u ./...

run: ## Run the main program
	go run cmd/htp/main.go

build: ## Build the binary
	go build -o bin/htp cmd/htp/main.go
