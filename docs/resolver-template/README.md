# Resolver Template

This directory contains a working Resolver based on the instructions
from the [developer howto in the docs](../how-to-write-a-resolver.md).

## Parameters

| Name   | Desccription                 | Example Value               |
|--------|------------------------------|-----------------------------|
| `url`  | The repository url.          | `https://example.com/repo/` |

## Using the template to start a new Resolver

You can use this as a template to quickly get a new Resolver up and
running with your own preferred storage backend.

To reuse the template, simply copy this entire subdirectory to a new
directory. The entire program is defined in
[`./cmd/myresolver/main.go`](./cmd/myresolver/main.go) and provides stub
implementations of all the methods defined by the [`framework.Resolver`
interface](../../pkg/resolver/framework/interface.go).

Once copied you'll need to run `go mod init` and `go mod tidy` at the root
of your project. We don't need this in `tektoncd/resolution` because this
submodule relies on the `go.mod` and `go.sum` defined at the root of the repo.

After your go module is initialized and dependencies tidied, update
`config/myresolver-deployment.yaml`. The `image` field of the container
will need to point to your new go module's name, with a `ko://` prefix.

## Deploying the Resolver

### Requirements

- A computer with
  [`kubectl`](https://kubernetes.io/docs/tasks/tools/#kubectl) and
  [`ko`](https://github.com/google/ko) installed.
- The `tekton-remote-resolution` namespace and `ResolutionRequest`
  controller installed. See [the getting started
  guide](./getting-started.md#step-3-install-tekton-resolution) for
  instructions.

### Install

1. Install the `"myresolver"` Resolver:

```bash
$ ko apply -f ./config/myresolver-deployment.yaml
```

### Testing it out

Try creating a `ResolutionRequest` targeting `"myresolver"` with no parameters:

```bash
$ cat <<EOF > rrtest.yaml
apiVersion: resolution.tekton.dev/v1alpha1
kind: ResolutionRequest
metadata:
  name: test-resolver-template
  labels:
    resolution.tekton.dev/type: myresolver
EOF

$ kubectl apply -f ./rrtest.yaml

$ kubectl get resolutionrequest -w test-resolver-template
```

You should shortly see the `ResolutionRequest` succeed and the content of
a hello-world `Pipeline` base64-encoded in the object's `status.data`
field.

## What's Supported?

- Just one hard-coded `Pipeline` for demonstration purposes.

---

Except as otherwise noted, the content of this page is licensed under the
[Creative Commons Attribution 4.0 License](https://creativecommons.org/licenses/by/4.0/),
and code samples are licensed under the
[Apache 2.0 License](https://www.apache.org/licenses/LICENSE-2.0).
