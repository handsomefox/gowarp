build: ## -- Build the project
	@echo Building...
	go build -o ./target/gowarp -ldflags "-s -w" ./cmd/http/main.go

vet: ## -- Run go vet
	go vet ./...

tidy: ## -- Run go mod tidy
	go mod tidy

lint: ## -- Run golangci-lint
	golangci-lint run ./...

fmt: ## -- Run gofumpt on the project
	gofumpt -l -w .

exec: ## -- Run the project
	@echo Running
	go run ./cmd/http/main.go

pre: ## -- Install the prerequisites
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.51.1
	go install mvdan.cc/gofumpt@latest
