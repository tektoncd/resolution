# Simple Git Resolver

## Resolver Type

This Resolver responds to type `git`.

## Parameters

| Param Name | Description                                                                  | Example Value                                               |
|------------|------------------------------------------------------------------------------|-------------------------------------------------------------|
| `url`      | URL of the repo to fetch.                                                    | `https://github.com/tektoncd/catalog.git`                   |
| `revision` | Git revision to checkout a file from. This can be commit SHA, branch or tag. | `aeb957601cf41c012be462827053a21a420befca` `main` `v0.38.2` |
| `pathInRepo` | Where to find the file in the repo.                                        | `/task/golang-build/0.3/golang-build.yaml`                  |

## Getting Started

### Requirements

- A cluster running [Tekton Pipelines from its main branch](https://github.com/tektoncd/pipeline)
  with the `alpha` feature gate enabled.
- `ko` installed.
- The `tekton-remote-resolution` namespace and `ResolutionRequest`
  controller installed. See [../README.md](../README.md).

### Install

1. Install the Git resolver:

```bash
$ ko apply -f ./gitresolver/config
```

## Configuration

This resolver uses a `ConfigMap` for its settings. See
[`./config/git-resolver-config.yaml`](./config/git-resolver-config.yaml)
for the name, namespace and defaults that the resolver ships with.

### Options

| Option Name | Description | Example Values |
|-------------|-------------|---------------|
| `fetch-timeout` | The maximum time any single git resolution may take. **Note**: a global maximum timeout of 1 minute is currently enforced on _all_ resolution requests. | `1m`, `2s`, `700ms` |

## Examples

### `PipelineRun`

```yaml
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  name: git-demo
spec:
  pipelineRef:
    resolver: git
    resource:
    - name: url
      value: https://github.com/tektoncd/catalog.git
    - name: revision
      value: main
    - name: pathInRepo
      value: pipeline/simple/0.1/simple.yaml
  params:
  - name: name
    value: Ranni
```

### `ResolutionRequest`

```bash
$ cat <<EOF > rrtest.yaml
apiVersion: resolution.tekton.dev/v1alpha1
kind: ResolutionRequest
metadata:
  name: fetch-catalog-task
  labels:
    resolution.tekton.dev/type: git
spec:
  params:
    url: https://github.com/tektoncd/catalog.git
    path: /task/golang-build/0.3/golang-build.yaml
EOF

$ kubectl apply -f ./rrtest.yaml

$ kubectl get resolutionrequest -w fetch-catalog-task
```

## What's Supported?

- At the moment the git resolver can only access public repositories.

---

Except as otherwise noted, the content of this page is licensed under the
[Creative Commons Attribution 4.0 License](https://creativecommons.org/licenses/by/4.0/),
and code samples are licensed under the
[Apache 2.0 License](https://www.apache.org/licenses/LICENSE-2.0).
