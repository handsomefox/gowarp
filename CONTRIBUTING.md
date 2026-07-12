# Contributing

Thanks for your interest in gowarp.

## Getting started

Install the tooling and vendor dependencies:

```shell
make pre
```

Copy `.env-example`, fill in the values (including the MongoDB database and collection names), and run from the repository root so the app can load assets from `./assets`:

```shell
make run_serve   # HTTP server
make run_cli     # CLI
```

## Before opening a pull request

Run the checks:

```shell
make fmt    # gofumpt
make vet    # go vet
make lint   # golangci-lint
```

Keep changes formatted with gofumpt and free of `go vet` and golangci-lint warnings. Describe what you changed and why, and call out anything that affects key generation, the CLI, the HTTP server, or database handling.
