# Development Guide

## Developer Setup

KubeIt uses [Task](https://taskfile.dev/) as the task runner for build automation and development workflows. To get started, ensure you have Task installed.

### Installing Task

You can install Task by following the instructions [here](https://taskfile.dev/installation/) or by doing:

```sh
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b ~/.local/bin
```

### Setting up the environment

`task setup`

## Build the project

`task build`

### Via Goreleaser

`goreleaser release --snapshot --skip-publish --rm-dist`

## Format the code

`task fmt`

More development guidelines and task automation details will be added soon.
