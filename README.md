# gowarp

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

An app to generate Cloudflare WARP+ keys, store them in a database, and
hand them out as you need. It ships both an HTTP server and a CLI.

## How to run

Makefile contains the commands for both building and running the project.
Before running, you should consider looking at `.env-example` file which
has an example of the environment variables that need to be set before
running the project.

To start the server, run:

```shell
make run_serve
```

To start the cli app, run:

```shell
make run_cli
```

The application expects the working directory to be the root of the project.
If it is not, it will error and exit on startup because of inability to load
assets from the `./assets` folder.

## Database

The project only supports working with MongoDB and thus expects you
to provide both the database name and collection name which will later
be used to store the generated keys.

## Testing

As of now, no tests are included, but later on I might add some.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Please report security-sensitive issues according to [SECURITY.md](SECURITY.md).

## License

Licensed under the [MIT License](LICENSE).
