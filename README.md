# Release Engineering Flow (Reflow)

![Status](https://img.shields.io/badge/status-active%20development-orange)

> :warning: **Reflow is in active development!**
> Expect rapid changes, breaking updates, and evolving features.

Reflow is SRE Release Engineering workflow interface and promotion tool. It provides a unified interface to the platform to define and promote services.

## Features

---

![Demo](docs/assets/reflow-demo.gif)

---

## Quick Start

Try out the Kubeit configurations in the [examples](./examples/)

1. Install Kubeit

   ```sh
   go install github.com/scorebet/reflow
   ```

1. Generate the labels to attach to your Docker container

   ```sh
   reflow_docker_labels=$(reflow generate docker-labels <reflow-resources-dir>)
   ```

1. Build your Docker container and add the Kubeit labels

   ```sh
   docker build $reflow_docker_labels -t docker.io/<namespace>:<tag>
   ```

1. Generate the Kubernetes manifest based on the labels attached to the Docker container

   ```sh
    reflow generate manifest docker.io/<namespace>:<tag>
   ```

1. The Kubernetes generated manifests are placed in `.reflow/generated/manifests.yaml`

---

## Roadmap

TBD

## CLI Docs

- [Reflow CLI](./docs/cli/reflow/reflow.md)
