# Bundles Resolver

## Parameters

| Param Name       | Description                                                                   | Example Value                                              |
|------------------|-------------------------------------------------------------------------------|------------------------------------------------------------|
| `namespace`      | The namespace of the service account to use                                   | `default`                                                  |
| `serviceAccount` | The name of the service account to use when constructing registry credentials | `default`                                                  |
| `bundle`         | The bundle url pointing at the image to fetch                                 | `gcr.io/tekton-releases/catalog/upstream/golang-build:0.1` |
| `name`           | The name of the resource to pull out of the bundle                            | `golang-build`                                             |
| `kind`           | The resource kind to pull out of the bundle                                   | `task`                                                     |

## Getting Started

### Requirements

- A cluster running this [in-progress pull request of Tekton Pipelines](https://github.com/tektoncd/pipeline/pull/4596)
  with the `alpha` feature gate enabled.
- `ko` installed.
- The `tekton-remote-resolution` namespace and `ResolutionRequest`
  controller installed. See [../README.md](../README.md).

### Install

1. Install the Bundles resolver:

```bash
$ ko apply -f ./bundleresolver/config
```

### Testing it out

Try creating a `ResolutionRequest` for a bundle:

```bash
$ cat <<EOF > rrtest.yaml
apiVersion: resolution.tekton.dev/v1alpha1
kind: ResolutionRequest
metadata:
  name: fetch-catalog-task
  labels:
    resolution.tekton.dev/type: bundle
spec:
  params:
    namespace: default
    serviceAccount: default
    bundle: gcr.io/tekton-releases/catalog/upstream/golang-build:0.1
    name: golang-build
    kind: task
EOF

$ kubectl apply -f ./rrtest.yaml

$ kubectl get resolutionrequest -w fetch-catalog-task
```

You should shortly see the `ResolutionRequest` succeed and the content of
the `golang-build.yaml` file base64-encoded in the object's `status.data`
field.

---

Except as otherwise noted, the content of this page is licensed under the
[Creative Commons Attribution 4.0 License](https://creativecommons.org/licenses/by/4.0/),
and code samples are licensed under the
[Apache 2.0 License](https://www.apache.org/licenses/LICENSE-2.0).
