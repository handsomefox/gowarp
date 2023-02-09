# gowarp

An app to generate Clodflare WARP+ keys, store them in a database and
get as you need.

## How to run

Makefile contains the commands for both building and running the project.
Before running, you should consider looking at `.env-example` file which
has an example of the environment variables that need to be set before
running the project.

To execute, run:

```bash
make exec
```

To build, run:

```bash
make build
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
