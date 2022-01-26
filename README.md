# Tekton Resolution

Pluggable resolution for Tekton resources (like `Tasks` and
`Pipelines`). Store and utilize Tekton resources from source,
from bundles, or from anywhere else.

Tekton Resolution is aiming for the following near-term goals:

- Pluggable. Allow integrations with Tekton Pipeline's resolution machinery
  without having to upstream changes to Tekton Pipelines.
- Configurable. Allow operators to choose which remote locations resources
  can be fetched from in their CI/CD clusters.
