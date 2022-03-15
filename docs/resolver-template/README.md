# Resolver Template

This directory contains a working Resolver based on the instructions
from the [developer howto in the docs](../how-to-write-a-resolver.md).

Copy this entire directory to quickly get started writing a new
Resolver. The entire program is defined in `./cmd/myresolver/main.go`.

## Getting Started

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

## Parameters

This Resolver has no parameters.

Except as otherwise noted, the content of this page is licensed under the
[Creative Commons Attribution 4.0 License](https://creativecommons.org/licenses/by/4.0/),
and code samples are licensed under the
[Apache 2.0 License](https://www.apache.org/licenses/LICENSE-2.0).
