# Simple Git Resolver

## Resolver Type

This Resolver responds to type `git`.

## Parameters

| Param Name | Description                                                                  | Example Value                                |
|------------|------------------------------------------------------------------------------|----------------------------------------------|
| `url`      | URL of the repo to fetch.                                                    | `https://github.com/tektoncd/catalog.git`    |
| `commit`   | git commit SHA to checkout a file from.                                      | `aeb957601cf41c012be462827053a21a420befca`   |
| `branch`   | The branch name to checkout a file from. Either this or commit but not both. | `main`                                       |
| `path`     | Where to find the file in the repo.                                          | `/task/golang-build/0.3/golang-build.yaml`   |

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

### Testing

Try creating a `ResolutionRequest` for a file in git:

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

You should shortly see the `ResolutionRequest` succeed and the content of
the `golang-build.yaml` file base64-encoded in the object's `status.data`
field.

### Example PipelineRun

Here's an example PipelineRun that pulls in a simple pipeline from a fork
of the Tekton Catalog:

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
      value: https://github.com/sbwsg/catalog.git
    - name: branch
      value: main
    - name: path
      value: pipeline/simple/0.1/simple.yaml
  params:
  - name: name
    value: Ranni
```

## What's Supported?

- At the moment the git resolver can only access public repositories.
- The git fetch must complete within 30 seconds. The `ResolutionRequest`
  object will be automatically failed after 60 seconds. Both of these
  timeouts need to be exposed for operator control via ConfigMap or
  similar but at the moment are just hard-coded.

---

Except as otherwise noted, the content of this page is licensed under the
[Creative Commons Attribution 4.0 License](https://creativecommons.org/licenses/by/4.0/),
and code samples are licensed under the
[Apache 2.0 License](https://www.apache.org/licenses/LICENSE-2.0).
