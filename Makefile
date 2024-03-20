GOFMT_FILES?=$$(find . -name '*.go')
GOFMT := "goimports"

fmt: ## Run gofmt for all .go files
	@$(GOFMT) -w $(GOFMT_FILES)

generate: fmt ## Generate proto & generate mock files
	go generate ./...

test: ## Run go test for whole project
	go test -v -race -coverprofile=coverage.out ./... && go tool cover -html=coverage.out && open coverage.out

coverage: ## Run go test for whole project
	go test -v -race -covermode=atomic -coverprofile=coverage.out ./...

lint: ## Run linter
	@golangci-lint run ./...

sonar: coverage ## Generate global code coverage report
	docker run --rm -v ".:/usr/src" -v "./coverage.out:/usr/src/coverage.out" sonarsource/sonar-scanner-cli

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

scan: ## Run security scan
	@docker buildx build --secret=id=netrc,src=$(HOME)/.netrc --output=type=docker,name=$(LOCAL_IMAGE),oci-mediatypes=true --tag=$(LOCAL_IMAGE):local -f Dockerfile.dev .
	trivy image $(LOCAL_IMAGE):local
	@docker rmi $(LOCAL_IMAGE):local
	trivy fs .