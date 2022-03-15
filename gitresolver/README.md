# Simple Git Resolver

## Getting Started

### Requirements

- A cluster running this [in-progress pull request of Tekton Pipelines](https://github.com/tektoncd/pipeline/pull/4596)
  with the `alpha` feature gate enabled.
- `ko` installed.
- The `tekton-remote-resolution` namespace and `ResolutionRequest`
  controller installed. See [../README.md](../README.md).

### Install

1. Install the Git resolver:

```bash
$ ko apply -f ./gitresolver/config
```

### Testing it out

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

## What's Supported?

- At the moment the git resolver can only access public repositories.
- The git fetch must complete within 30 seconds. The `ResolutionRequest`
  object will be automatically failed after 60 seconds. Both of these
  timeouts need to be exposed for operator control via ConfigMap or
  similar but at the moment are just hard-coded.

## Parameters


| Param Name | Description                                                                  | Example Value                                |
|------------|------------------------------------------------------------------------------|----------------------------------------------|
| `url`      | URL of the repo to fetch.                                                    | `https://github.com/tektoncd/catalog.git`    |
| `commit`   | git commit SHA to checkout a file from.                                      | `aeb957601cf41c012be462827053a21a420befca`   |
| `branch`   | The branch name to checkout a file from. Either this or commit but not both. | `main`                                       |
| `path`     | Where to find the file in the repo.                                          | `/task/golang-build/0.3/golang-build.yaml`   |
