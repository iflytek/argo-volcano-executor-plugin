package controller

import (
	"argo-volcano-executor-plugin/pkg/utils/jsonUtil"
	"encoding/json"
	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	executorplugins "github.com/argoproj/argo-workflows/v3/pkg/plugins/executor"
	"github.com/gin-gonic/gin"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"net/http"
	"time"
)

type K8sController struct {
	KubeClient *kubernetes.Clientset
}

type K8sJobBody struct {
	Job *batchv1.Job `json:"job"`
}

type K8sPluginBody struct {
	JobBody *K8sJobBody `json:"k8s"`
}

func (ct *K8sController) ExecuteK8sJob(ctx *gin.Context) {
	c := &executorplugins.ExecuteTemplateArgs{}
	err := ctx.BindJSON(&c)
	if err != nil {
		klog.Error(err)
		return
	}
	// Get Job First
	inputBody := &K8sPluginBody{
		JobBody: &K8sJobBody{
			Job: &batchv1.Job{},
		},
	}
	pluginJson, _ := c.Template.Plugin.MarshalJSON()
	var object interface{}
	err = json.Unmarshal(pluginJson, &object)
	if err != nil {
		klog.Error(err)
		ct.Response404(ctx)
		return

	}
	klog.Info("Receive: ", string(pluginJson))
	err = jsonUtil.UnmarshalFromMap(object, inputBody)
	//err = json.Unmarshal(pluginJson, inputBody)
	if err != nil {
		klog.Error(err)
		ct.Response404(ctx)
		return

	}
	job := inputBody.JobBody.Job
	if job.Name == "" {
		job.ObjectMeta.Name = c.Workflow.ObjectMeta.Name

	}
	if job.ObjectMeta.Namespace == "" {
		job.Namespace = "default"
	}

	var exists = false

	// 1. query job exists
	existsJob, err := ct.KubeClient.BatchV1().Jobs(job.Namespace).Get(ctx, job.Name, metav1.GetOptions{})
	if err != nil {
		exists = false
	} else {
		exists = true
	}

	// 2. found and return
	if exists {
		klog.Info("# found exists Volcano Job: ", job.Name, "returning Status...", job.Status)
		ct.ResponseK8sJob(ctx, existsJob)
		return
	}

	// 3.Label keys with workflow Name
	InjectK8sJobWithWorkflowName(job, c.Workflow.ObjectMeta.Name)

	newJob, err := ct.KubeClient.BatchV1().Jobs(job.Namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		klog.Error("### " + err.Error())
		ct.ResponseMsg(ctx, wfv1.NodeFailed, err.Error())
		return
	}

	ct.ResponseCreated(ctx, newJob)

}

func (ct *K8sController) ResponseCreated(ctx *gin.Context, job *batchv1.Job) {

	ctx.JSON(http.StatusOK, &executorplugins.ExecuteTemplateReply{
		Node: &wfv1.NodeResult{
			Phase:   wfv1.NodePending,
			Message: job.Status.String(),
			Outputs: nil,
		},
		Requeue: &metav1.Duration{
			Duration: 10 * time.Second,
		},
	})
}

func (ct *K8sController) ResponseMsg(ctx *gin.Context, status wfv1.NodePhase, msg string) {
	ctx.JSON(http.StatusOK, &executorplugins.ExecuteTemplateReply{
		Node: &wfv1.NodeResult{
			Phase:   status,
			Message: msg,
			Outputs: nil,
		},
	})
}

func (ct *K8sController) ResponseK8sJob(ctx *gin.Context, job *batchv1.Job) {
	var requeue *metav1.Duration
	var status wfv1.NodePhase

	if status == wfv1.NodeRunning || status == wfv1.NodePending {
		requeue = &metav1.Duration{
			Duration: 10 * time.Second,
		}
	} else {
		requeue = nil
	}
	for _, condition := range job.Status.Conditions {
		if condition.Type == batchv1.JobComplete && condition.Status == corev1.ConditionTrue {
			// Job 完成
			status = wfv1.NodeSucceeded
		} else if condition.Type == batchv1.JobFailed && condition.Status == corev1.ConditionTrue {
			// Job 失败
			status = wfv1.NodeFailed
		} else {
			status = wfv1.NodeRunning
			requeue = &metav1.Duration{
				Duration: 10 * time.Second,
			}
		}
	}

	succeed := job.Status.Succeeded
	// not sure here
	Total := job.Status.Failed + job.Status.Succeeded + job.Status.Active
	progress, _ := wfv1.NewProgress(int64(succeed), int64(Total))
	klog.Info("### K8s Job Phase: ", status)

	ctx.JSON(http.StatusOK, &executorplugins.ExecuteTemplateReply{
		Node: &wfv1.NodeResult{
			Phase:    status,
			Message:  job.Status.String(),
			Outputs:  nil,
			Progress: progress,
		},
		Requeue: requeue,
	})
}

func (ct *K8sController) Response404(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotFound)

}

func InjectK8sJobWithWorkflowName(job *batchv1.Job, workflowName string) {
	if job.Spec.Template.ObjectMeta.Labels != nil {
		job.Spec.Template.ObjectMeta.Labels[LabelKeyWorkflow] = workflowName
	} else {
		job.Spec.Template.ObjectMeta.Labels = map[string]string{
			LabelKeyWorkflow: workflowName,
		}

	}
	klog.Info("Injecting Labels with workflow name:", workflowName)
}
