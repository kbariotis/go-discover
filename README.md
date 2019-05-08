# Discover

## Usage

```sh
export GITHUB_TOKEN=...
make run
```

__Available env vars:__

* `GITHUB_TOKEN` - github token, __required__
* `LOG_LEVEL` - log level: `error`, `info`, `debug`, `trace`; defaults to `info`
* `QUEUE_STORE_DIR` - path for dqueue persistnce; defaults to `~/go-discover`

## Development

```sh
make tools
make deps
make lint
make test
```

__Available make targets:__

* `make deps` - installs dependencies
* `make tools` - installs required tools under `./bin`
* `make lint` - lints package using `./bin/golangci-lint`
* `make build` - builds `cmd/discover` as `./bin/discover`
* `make run` - builds and runs `cmd/discover`
* `make test` - tests package
* `make clean` - removes temp files