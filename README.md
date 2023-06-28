# todo-list

TODO list app for Golang Syd June 2023 Interactive Session

## Pre-requisites

Before the Session:

* [Docker](https://www.docker.com/products/docker-desktop/) is installed and running
* Golang v1.20 SDK is installed and working

This project has been developed on macOS; while it is expected to work elsewhere, it has not been tested.

To make the session as smooth as possible, ideally you also have the project ready to use:

* The repository has been cloned: `git clone git@github.org:dackroyd/todo-list.git`
* Docker images used by `docker-compose.yaml` have been fetched & built. From the root of the repo:
    * `docker compose pull`
    * `docker compose --profile frontend build simulate-ui`
    * Optional: only if you intend to build/run the app in Docker `docker compose --profile backend build api`
* Go module dependencies have been fetched `cd backend && go mod download`
