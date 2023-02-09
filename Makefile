# Build all targets.
all: tidy cli serve
	@echo Built all targets...

# Build the CLI
cli:
	@echo Building CLI...
	go build -o ./target/gowarp-cli -ldflags "-s -w" ./cmd/cli/main.go

# Build the server
serve:
	@echo Building server...
	go build -o ./target/gowarp-serve -ldflags "-s -w" ./cmd/http/main.go

# Remove build artifacts
clean:
	@echo Removing build artifacts...
	rm -r ./target

# Run the cli application
run_cli: cli
	@echo Starting CLI...
	./target/gowarp-cli

# Start the server
run_serve: serve
	@echo Starting the server...
	./target/gowarp-serve

# Run go vet
vet:
	@echo Running go vet...
	go vet ./...

# Run go mod tidy
tidy:
	@echo Running go mod tidy...
	go mod tidy

# Run golangci-lint
lint:
	@echo Running golangci-lint...
	golangci-lint run ./...

# Run gofumpt on the project
fmt:
	@echo Running gofumpt...
	gofumpt -l -w .

# Install the prerequisites
pre: tidy
	@echo Installing prerequisites...
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.51.1
	go install mvdan.cc/gofumpt@latest
	@echo Vendoring dependencies...
	go mod vendor
