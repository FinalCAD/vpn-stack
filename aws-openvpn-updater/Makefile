PROJECT_NAME := "aws-openvpn-updater"
GO_DEST := "build"
GOMAIN := "./cmd/aws-openvpn-updater/main.go"

default: build

dep: fmt ## Get the dependencies
	@go mod tidy

build: dep ## Build
	@go build -o $(GO_DEST)/$(PROJECT_NAME) $(GOMAIN)

clean: ## Remove builds
	@rm -f ./$(GO_DEST)/*

run: build ## Run project
	@go run $(GOMAIN) -env sandbox

debug: build ## Run project in debug
	@go run $(GOMAIN) -debug -env sandbox

fmt:
	@go fmt ./...

test: build ## Run test
	@go test -v ./...

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
