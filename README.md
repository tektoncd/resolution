# Tekton Resolution

Pluggable resolution for Tekton resources (like `Tasks` and
`Pipelines`). Store and utilize Tekton resources from git,
from oci registries, or from anywhere else.

Tekton Resolution is aiming for the following near-term goals:

- Pluggable. Allow integrations with Tekton Pipeline's resolution machinery
  without having to upstream changes to Tekton Pipelines.
- Configurable. Allow operators to choose which remote locations resources
  can be fetched from in their CI/CD clusters.

## Getting Started

**Warning: Tekton Resolution is under development and currently only
suitable for those interested in testing the project at its bleeding
edge. Breakages to the API, types, libraries and resolvers are unavoidable
at this stage.**

### Requirements

- A cluster runnning the very latest version of Tekton Pipelines with
  the `alpha` feature gate enabled.
- `ko` installed.

### Install

Out of the box Tekton Resolution provides a simple Git resolver that can
fetch files from public git repositories.

1. Create the `tekton-remote-resolution` namespace and install
the `ResourceRequest` controller from the root of this repo:

```bash
$ ko apply -f ./config
```

2. Install [the Git resolver](./gitresolver/README.md).
