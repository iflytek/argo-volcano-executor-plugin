# Argo volcano job plugin


## 背景

### 官方描述背景:

As  [issue-7860](https://github.com/argoproj/argo-workflows/issues/7860) says:

* It is expensive as it requires one pod per resource.
  当前argo创建 K8s Resource需要单独启动1个pod来管理resource的生命周期代价有些高。

* It only supports a single resource.

* Resource manifests must come from the workflow.
  Resource定义必须来源于workflow定义中

* Resources are strings, not structured YAML

### 补充背景with volcano

当前和volcano CRD JOB结合时有以下问题:

* If volcano job is `Pending` status because of some reasons such as lack of resource, the
  workflow status is Running (with a pod running).

* If volcano job status is success or failed, the success condition  or failedCondition is not
  very flexible in argo.

### Proposal

So I have make a proposal:


A specified Plugin for Volcano Job



