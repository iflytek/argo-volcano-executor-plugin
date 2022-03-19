# Argo volcano job plugin
## About The Project
## 官方描述背景

As  [issue-7860](https://github.com/argoproj/argo-workflows/issues/7860) says:

* It is expensive as it requires one pod per resource.
  当前argo创建 K8s Resource需要单独启动1个pod来管理resource的生命周期代价有些高。

* It only supports a single resource.

* Resource manifests must come from the workflow.
  Resource定义必须来源于workflow定义中

* Resources are strings, not structured YAML

### Why this plugin?

当前和volcano CRD JOB结合时有以下问题:

* If volcano job is `Pending` status because of some reasons such as lack of resource, the
  workflow status is Running (with a pod running).

* If volcano job status is success or failed, the success condition  or failedCondition is not
  very flexible in argo.

So I have make this project:
A specified Plugin for Volcano Job

### Built with

Open Source software stands on the shoulders of giants. It wouldn't have been possible to build this tool with very little extra work without the authors of these lovely projects below.

* [Gin](https://github.com/gin-gonic/gin) Golang HTTP FrameWork
* [Volcano API]


## Getting Started

Todo here

### Prerequisites

* This guide assumes you have a working Argo Workflows installation with v3.3.0 or newer.
* You will need to install the [Argo CLI](https://argoproj.github.io/argo-workflows/cli/) with v3.3.0 or newer.
* `kubectl` must be available and configured to access the Kubernetes cluster where Argo Workflows is installed.

### Installation

1. Clone the repository and change to the [`argo-plugin/`](argo-plugin/) directory:

   ```shell
   git https://github.com/xfyun/argo-volcano-executor-plugin
   cd argo-volcano-executor-plugin
   ```

1. Build the plugin ConfigMap:

   ```shell
   argo executor-plugin build .
   ```

1. Register the plugin with Argo in your cluster:

   Ensure to specify `--namespace` if you didn't install Argo in the default namespace.

   ```shell
   kubectl apply -f volcano-executor-plugin-configmap.yaml
   ```

## Usage

Now that the plugin is registered in Argo, you can run workflow steps as Wasm modules by simply calling the `wasm` plugin:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  generateName: hello-
spec:
  automountServiceAccountToken: true
  entrypoint: main
  templates:
    - name: main
      executor:
        serviceAccountName: default
      plugin:
        volcano:
          job:
            minAvailable: 3
            schedulerName: volcano
            plugins:
              env: []
              svc: []
            queue: default
            policies:
              - event: PodEvicted
                action: RestartJob
              - event: TaskCompleted
                action: CompleteJob
            tasks:
              - replicas: 1
                name: ps
                template:
                  spec:
                    containers:
                      - command:
                          - sh
                          - -c
                          - |
                            PS_HOST=`cat /etc/volcano/ps.host | sed 's/$/&:2222/g' | sed 's/^/"/;s/$/"/' | tr "\n" ","`;
                            WORKER_HOST=`cat /etc/volcano/worker.host | sed 's/$/&:2222/g' | sed 's/^/"/;s/$/"/' | tr "\n" ","`;
                            export TF_CONFIG={\"cluster\":{\"ps\":[${PS_HOST}],\"worker\":[${WORKER_HOST}]},\"task\":{\"type\":\"ps\",\"index\":${VK_TASK_INDEX}},\"environment\":\"cloud\"};
                            python /var/tf_dist_mnist/dist_mnist.py
                        image: volcanosh/dist-mnist-tf-example:0.0.1
                        name: tensorflow
                        ports:
                          - containerPort: 2222
                            name: tfjob-port
                        resources: {}
                    restartPolicy: Never
              - replicas: 2
                name: worker
                policies:
                  - event: TaskCompleted
                    action: CompleteJob
                template:
                  spec:
                    containers:
                      - command:
                          - sh
                          - -c
                          - |
                            PS_HOST=`cat /etc/volcano/ps.host | sed 's/$/&:2222/g' | sed 's/^/"/;s/$/"/' | tr "\n" ","`;
                            WORKER_HOST=`cat /etc/volcano/worker.host | sed 's/$/&:2222/g' | sed 's/^/"/;s/$/"/' | tr "\n" ","`;
                            export TF_CONFIG={\"cluster\":{\"ps\":[${PS_HOST}],\"worker\":[${WORKER_HOST}]},\"task\":{\"type\":\"worker\",\"index\":${VK_TASK_INDEX}},\"environment\":\"cloud\"};
                            python /var/tf_dist_mnist/dist_mnist.py
                        image: volcanosh/dist-mnist-tf-example:0.0.1
                        name: tensorflow
                        ports:
                          - containerPort: 2222
                            name: tfjob-port
                        resources: {}
                    restartPolicy: Never
```

The `volcano` template will produce vcjob that you can use command `kubect get vcjob ` to browse them .



## Roadmap

- [ ] Support config service account


## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

If you have a suggestion that would make this better, please fork the repo and create a pull request. You can also simply open an issue with the tag "enhancement".
Don't forget to give the project a star! Thanks again!

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request


## Contact

Project Link: [https://github.com/xfyun/argo-volcano-executor-plugin](https://github.com/xfyun/argo-volcano-executor-plugin)

