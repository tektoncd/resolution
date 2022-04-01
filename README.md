# Tekton Resolution

**Warning: Tekton Resolution is under development and currently only
suitable for those interested in testing the project at its bleeding
edge. Breakages to the API, types, libraries and resolvers are unavoidable
at this stage.**

Pluggable resolution for Tekton resources (like `Tasks` and
`Pipelines`). Store and utilize Tekton resources from git,
from oci registries, or from anywhere else.

Tekton Resolution is aiming for the following near-term goals:

- Pluggable. Allow integrations with Tekton Pipeline's resolution machinery
  without having to upstream changes to Tekton Pipelines.
- Configurable. Allow operators to choose which remote locations resources
  can be fetched from in their CI/CD clusters.

## Getting Started

### Requirements

- A cluster running this [in-progress pull request of Tekton Pipelines](https://github.com/tektoncd/pipeline/pull/4596)
  with the `alpha` feature gate enabled.
- `ko` installed.

### Install

Out of the box Tekton Resolution provides a simple Git resolver that can
fetch files from public git repositories.

1. Create the `tekton-remote-resolution` namespace and install
the `ResolutionRequest` controller from the root of this repo:

```bash
$ ko apply -f ./config
```

2. [Install a resolver](#resolvers) or [get started writing your
   own](./docs/how-to-write-a-resolver.md).

## Resolvers

| Name                                                        | Description                                                                     | Status    |
|-------------------------------------------------------------|---------------------------------------------------------------------------------|-----------|
| [`Bundle`](./bundleresolver)                                | Returns entries from oci bundles                                                | Pre-Alpha |
| [`Git`](./gitresolver)                                      | Returns files from git repos                                                    | Pre-Alpha |
| [`ClusterScoped`](https://github.com/sbwsg/clusterresolver) | Share a single set of tasks and pipelines across all namespaces in your cluster | Pre-Alpha |

---

Except as otherwise noted, the content of this page is licensed under the
[Creative Commons Attribution 4.0 License](https://creativecommons.org/licenses/by/4.0/),
and code samples are licensed under the
[Apache 2.0 License](https://www.apache.org/licenses/LICENSE-2.0).
