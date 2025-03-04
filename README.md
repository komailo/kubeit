# Kubeit 🏗️🚀

![Status](https://img.shields.io/badge/status-active%20development-orange)

> :warning: **Kubeit is in active development!**  
> Expect rapid changes, breaking updates, and evolving features.

Kubeit is a deployment automation tool that simplifies Kubernetes configuration for service teams and platform engineers.

It eliminates the need for manually managing Kubernetes manifests or Helm charts by transforming minimal Kubeit configuration into fully functional Kubernetes objects.

By embedding deployment configuration inside the container, Kubeit enables self-contained deployable containers that generate their Kubernetes manifests dynamically at runtime—eliminating the need for a separate Kubernetes deployment. This reduces the number of artifacts required for deployment and ensures that Kubernetes configurations are always in sync with the application—eliminating the risk of missing or outdated dependencies.

Why Choose Kubeit?

- Minimal YAML, Maximum Simplicity – Define your infrastructure in a straightforward format, reducing complexity.

- Fewer Deployment Artifacts – No need to manage separate Kubernetes manifests or Helm charts — Kubeit keeps everything self-contained.

- Always In Sync – Kubernetes configurations are generated alongside the container, ensuring no missing dependencies between application and infrastructure.

- Extensible & Flexible – Supports Helm, CRDs, and custom templates to fit diverse deployment needs.

- Seamless CI/CD Integration – Works with existing pipelines and developer workflows for smooth deployments.

- Portable & Consistent – Keep your deployment settings alongside your container for better traceability and portability.

## Roadmap

### Active Feature Development (MVP)

- :heavy_check_mark: Transform Kubeit configuration into Kubernetes objects automatically.

- :heavy_check_mark: Embed Kubeit configuration directly into the container image for streamlined deployment.

### Next Phase

- Enable deployments via tools like ArgoCD to facilitate GitOps workflows.

### Future Enhancements

- Introduce rollout strategies for more controlled and seamless deployments.

- Expand integrations with additional deployment and orchestration tools.

- Improve customization options for advanced use cases.

## CLI Docs

- [Kubeit CLI](./docs/cli/kubeit/kubeit.md)
