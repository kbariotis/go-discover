# Discover

> A weekly GitHub newsletter tailored on you. Discover what your peers are doing on GitHub.

Discover is monitoring the people you follow on GitHub and sends you a weekly newsletter with that activity. Discover projects that your friends are working on, new repositories that they recently discovered, issues and releases that you probably missed.

It consists of three parts.

__Crawler__

The crawler is the part that monitors GitHub activity. It constantly crawls the activity of the people you follow and feeds our Graph DB. Based on that, then, we are able to export data based on your connections.

__API__

The API is serving our website. Currently is serving HTML as well. (@TODO move HTML/landing page out of the API in its own container)

__Extraction (Better name pending?!)__

Extraction is the part that queries our GraphDB, prepares the email template and sends it out as a weekly newsletter.

## Technologies

This project is using:

* Written in Go
* Neo4j for our GraphDB
* docker/docker-compose
* PostgreSQL for everything relational
* Redis for cache
* Mailgun for emails
* GitHub duh?

## Configuration

| Variable | Description | Required | Default |
| --- | --- | --- | --- |
| `GITHUB_TOKEN` | GitHub token for the crawler | yes | |
| `LOG_LEVEL` | Log level: `error`, `info`, `debug`, `trace` | no | `info` |
| `QUEUE_STORE_DIR` | | no | `~/go-discover` |
| `SUGGESTION_STORE_TYPE` | | no | postgres |
| `SUGGESTION_STORE_HOST` | | no | postgres |
| `SUGGESTION_STORE_PORT` | | no | 5432 |
| `SUGGESTION_STORE_USER` | | no | postgres |
| `SUGGESTION_STORE_DB` | | no | discover |
| `SUGGESTION_STORE_PWD` | | yes |
| `NEO4J_HOST` | | no | http://localhost:7474/db/data |
| `REDIS_HOST` | | no | localhost:6379 |
| `API_BIND_ADDRESS` | | no | 0.0.0.0:8080 |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth secret | yes | |
| `GITHUB_CLIENT_ID` | GitHub OAuth ID | yes | |
| `GITHUB_CALLBACK_URL` | GitHub OAuth callback URL | no | http://localhost:8080/github/callback |
| `MAILGUN_DOMAIN` | | yes | |
| `MAILGUN_APIKEY` | | yes | |
| `MAIL_SENDER_ADDRESS` | | yes | |
| `LOCK_USER_DURATION` | | no | 12h |
| `LOCK_REPOSITORY_DURATION` | | no | 24h | |

## Development

For local development we use `docker-compose`.

**Make sure that you have a valid `.env` file in place.**

Start everything up:

`> docker-compose up`

From there you can go to `localhost:8080` to login with GitHub. Next time `crawler` runs, will pick up your registered user and start crawling in your behalf.

__Available make targets:__

* `make deps` - installs dependencies
* `make tools` - installs required tools under `./bin`
* `make lint` - lints package using `./bin/golangci-lint`
* `make build-api` - builds `cmd/api` as `./bin/api` _(used for building for production only)_
* `make build-crawler` - builds `cmd/crawler` as `./bin/crawler` _(used for building for production only)_
* `make build-extraction` - builds `cmd/extraction` as `./bin/extraction` _(used for building for production only)_
* `make run-api` - builds and runs `cmd/api`
* `make run-crawler` - builds and runs `cmd/crawler`
* `make run-extraction` - builds and runs `cmd/extraction`
* `make test` - tests package
* `make clean` - removes temp files

## Going to production

Discover is running on a Kubernetes cluster and the configuration files live under [infrastructure](https://github.com/kbariotis/go-discover/tree/master/infrastructure).

To build any container of the services use the corresponding Dockerfile under root.

Example building `api`:

`> docker build -t kostarelo/discover-api -f Dockerfile.api .`

and push to the registry. (@TODO add service to automatically fetch new versions of the containers running.)

## Contribute

Want to contribute to this project?

We are absolutely welcoming contribution to this project. Take a look at our open [issues](https://github.com/kbariotis/go-discover) and open a PR if you see one you can help.

Feel free to get in touch with us to discuss more about this project and how you can get involved. We love hearing from people!

## Creators

| Name | Twitter |
| --- | --- |
| Kostas Bariotis | [@kbariotis](https://twitter.com/kbariotis) |
| George Antoniadis | [@geoah](https://twitter.com/geoah) |
