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

- A cluster running the [Tekton Pipelines from its main branch](https://github.com/tektoncd/pipeline)
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

Resolvers do the heavy lifting fetching tekton resources from remote places (like repos, registries, etc...). These are the resolvers that are currently implemented. Once a Resolver is installed in your Tekton cluster all users in that cluster can start making use of it.

| Name                                                        | Description                                                                     | Status    |
|-------------------------------------------------------------|---------------------------------------------------------------------------------|-----------|
| [`Bundle`](./bundleresolver)                                | Returns entries from oci bundles                                                | Alpha |
| [`Git`](./gitresolver)                                      | Returns files from git repos                                                    | Alpha |
| [`ClusterScoped`](https://github.com/sbwsg/clusterresolver) | Shares a single set of tasks and pipelines across all namespaces in your cluster | Alpha |

Want to integrate with a remote location that isn't listed here? [Write a new resolver](./docs/how-to-write-a-resolver.md) or [post an issue requesting one](https://github.com/tektoncd/resolution/issues/new?assignees=&labels=kind%2Ffeature&template=feature-request.md).

---

Except as otherwise noted, the content of this page is licensed under the
[Creative Commons Attribution 4.0 License](https://creativecommons.org/licenses/by/4.0/),
and code samples are licensed under the
[Apache 2.0 License](https://www.apache.org/licenses/LICENSE-2.0).
