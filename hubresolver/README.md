# Hub Resolver

Use resolver type `hub`.

## Parameters

| Param Name       | Description                                                                   | Example Value                                              |
|------------------|-------------------------------------------------------------------------------|------------------------------------------------------------|
| `kind`           | Either `task` or `pipeline`                                                   | `task`                                                     |
| `name`           | The name of the task or pipeline to fetch from the hub                        | `golang-build`                                             |
| `version`        | Version of task or pipeline to pull in from hub. Wrap the number in quotes!   | `"0.5"`                                                    |

## Getting Started

### Requirements

See the [getting started
instructions](https://github.com/tektoncd/resolution/tree/main/docs/getting-started.md)
in the Tekton Resolution repo.

### Install

1. Install the Hub resolver:

```bash
$ ko apply -f ./config
```

### Configuring the Hub API endpoint

By default this resolver will hit the public hub api at https://hub.tekton.dev/
but you can configure your own (for example to use a private hub
instance) by setting the `HUB_API` environment variable in
`config/hubresolver-deployment.yaml`. Example:

```yaml
env
- name: HUB_API
  value: "https://api.hub.tekton.dev/"
```

### Testing it out

Try creating a `ResolutionRequest` for a hub entry:

```bash
$ cat <<EOF > rrtest.yaml
apiVersion: resolution.tekton.dev/v1alpha1
kind: ResolutionRequest
metadata:
  name: fetch-hub-entry
  labels:
    resolution.tekton.dev/type: hub
spec:
  params:
    kind: task
    name: git-clone
    version: "0.5"
    kind: task
EOF

$ kubectl apply -f ./rrtest.yaml

$ kubectl get resolutionrequest -w fetch-hub-entry
```

You should shortly see the `ResolutionRequest` succeed and the content of
the `git-clone` yaml base64-encoded in the object's `status.data`
field.

### Example PipelineRun

```yaml
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  name: hub-demo
spec:
  pipelineRef:
    resolver: hub
    resource:
    - name: kind
      value: pipeline
    - name: name
      value: buildpacks
    - name: version
      value: "0.1"
  # Note: the buildpacks pipeline requires parameters.
  # Resolution of the pipeline will succeed but the PipelineRun
  # overall will not succeed without those parameters.
```

---

Except as otherwise noted, the content of this page is licensed under the
[Creative Commons Attribution 4.0 License](https://creativecommons.org/licenses/by/4.0/),
and code samples are licensed under the
[Apache 2.0 License](https://www.apache.org/licenses/LICENSE-2.0).
