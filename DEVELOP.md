# Development Guide

## Build & Development Workflow

KubeIt uses [Task](https://taskfile.dev/) as the task runner for build automation and development workflows. To get started, ensure you have Task installed.

### Installing Task

You can install Task by following the instructions [here](https://taskfile.dev/installation/) or by doing:

```sh
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b ~/.local/bin
```

## Build the project

task build

## Format the code

task fmt

More development guidelines and task automation details will be added soon.

### Installing golang-ci lint

`go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.5`
