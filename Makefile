APP = locals

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

deps: ## Ensure dependencies
	go mod tidy
	go mod download

lint: ## Run lint
	golangci-lint run

test: ## Test and print coverage report
	mkdir reports || true
	go test -v -cover -coverprofile reports/coverage-report.out ./...

clean: ## Clean up build folder
	rm build/*/*

compile: ## Build 
	GO_ENABLED=1 go build -mod=mod -tags musl -o build/${APP} -a .

run: build ## Run
	$(GOBIN)/$(APP)