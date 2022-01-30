# Simple Git Resolver

## Getting Started

### Requirements

- A cluster runnning the very latest version of Tekton Pipelines with
  the `alpha` feature gate enabled.
- `ko` installed.
- The `tekton-remote-resolution` namespace and `ResourceRequest`
  controller installed. See [../README.md](../README.md).

### Install

1. Install the Git resolver:

```bash
$ ko apply -f ./gitresolver/config
```

### Testing it out

Try creating a `ResourceRequest` for a file in git:

```bash
$ cat <<EOF > rrtest.yaml
apiVersion: resolution.tekton.dev/v1alpha1
kind: ResourceRequest
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

$ kubectl get resourcerequest -o yaml -w fetch-catalog-task
```

You should shortly see the `ResourceRequest` resolved successfully with
the base64-encoded contents of the `golang-build.yaml` file from the
Tekton Catalog.

## What's Supported?

- At the moment the git resolver can only access public repositories.
- The git fetching operation must complete within 30 seconds.
