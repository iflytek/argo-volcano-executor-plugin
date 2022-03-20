package controller

import (
	"encoding/json"
	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	executorplugins "github.com/argoproj/argo-workflows/v3/pkg/plugins/executor"
	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"net/http"
	"time"
	batch "volcano.sh/apis/pkg/apis/batch/v1alpha1"
	"volcano.sh/apis/pkg/client/clientset/versioned"
)

const (
	LabelKeyWorkflow string = "workflows.argoproj.io/workflow"
)

type Controller struct {
	VcClient   *versioned.Clientset
	KubeClient *kubernetes.Clientset
}

type JobBody struct {
	JobSpec *batch.JobSpec `json:"job"`
}

type VolcanoPluginBody struct {
	JobBody *JobBody `json:"volcano"`
}

func (ct *Controller) ExecuteVolcanoJob(ctx *gin.Context) {
	klog.Info("#####starting......")
	c := &executorplugins.ExecuteTemplateArgs{}
	err := ctx.BindJSON(&c)
	if err != nil {
		klog.Error(err)
		return
	}
	inputBody := &VolcanoPluginBody{
		JobBody: &JobBody{
			JobSpec: &batch.JobSpec{},
		},
	}
	pluginJson, _ := c.Template.Plugin.MarshalJSON()
	err = json.Unmarshal(pluginJson, inputBody)
	if err != nil {
		klog.Error(err)
		ct.Response404(ctx)
		return

	}
	var msg string
	jobSpec := inputBody.JobBody.JobSpec
	if jobSpec.MinAvailable < 0 {
		msg = "job 'minAvailable' must be >= 0."
		klog.Error(msg)
		ct.ResponseMsg(ctx, wfv1.NodeFailed, msg)
		return
	}

	if jobSpec.MaxRetry < 0 {
		msg = "'maxRetry' cannot be less than zero."
		klog.Error(msg)
		ct.ResponseMsg(ctx, wfv1.NodeFailed, msg)
		return
	}
	job := &batch.Job{
		Spec: *jobSpec,
		ObjectMeta: metav1.ObjectMeta{
			Name: c.Workflow.ObjectMeta.Name,
		},
	}
	if job.Namespace == "" {
		job.Namespace = "default"
	}

	var exists = false

	// 1. query job exists
	existsJob, err := ct.VcClient.BatchV1alpha1().Jobs(job.Namespace).Get(ctx, job.Name, metav1.GetOptions{})
	if err != nil {
		exists = false
	} else {
		exists = true
	}

	// 2. found and return
	if exists {
		klog.Info("# found exists Volcano Job: ", job.Name, "Returning Status...")
		ct.ResponseVcJob(ctx, existsJob)
		return
	}

	// 3.Label keys with workflow Name
	InjectVcJobWithWorkflowName(job, c.Workflow.ObjectMeta.Name)

	newJob, err := ct.VcClient.BatchV1alpha1().Jobs(job.Namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		klog.Error("### " + err.Error())
		ct.ResponseMsg(ctx, wfv1.NodeFailed, err.Error())
		return
	}

	if newJob.Spec.Queue == "" {
		newJob.Spec.Queue = "default"
	}
	ct.ResponseCreated(ctx)

}

func (ct *Controller) ResponseCreated(ctx *gin.Context) {

	ctx.JSON(http.StatusOK, &executorplugins.ExecuteTemplateReply{
		Node: &wfv1.NodeResult{
			Phase:   wfv1.NodePending,
			Message: "",
			Outputs: nil,
		},
		Requeue: &metav1.Duration{
			Duration: 10 * time.Second,
		},
	})
}

func (ct *Controller) ResponseMsg(ctx *gin.Context, status wfv1.NodePhase, msg string) {
	ctx.JSON(http.StatusOK, &executorplugins.ExecuteTemplateReply{
		Node: &wfv1.NodeResult{
			Phase:   status,
			Message: msg,
			Outputs: nil,
		},
	})
}

func (ct *Controller) ResponseVcJob(ctx *gin.Context, job *batch.Job) {
	jobPhase := &job.Status.State.Phase
	var status wfv1.NodePhase
	switch *jobPhase {
	case batch.Running:
		status = wfv1.NodeRunning
	case batch.Aborted:
		status = wfv1.NodeError
	case batch.Completed:
		status = wfv1.NodeSucceeded
	case batch.Pending:
		status = wfv1.NodePending
	case batch.Failed:
		status = wfv1.NodeFailed
	default:
		status = wfv1.NodeRunning
	}
	var requeue *metav1.Duration
	if status == wfv1.NodeRunning || status == wfv1.NodePending {
		requeue = &metav1.Duration{
			Duration: 10 * time.Second,
		}
	} else {
		requeue = nil
	}
	succeed := job.Status.Succeeded
	// not sure here
	Total := job.Status.Failed + job.Status.Succeeded + job.Status.Running
	progress, _ := wfv1.NewProgress(int64(succeed), int64(Total))

	ctx.JSON(http.StatusOK, &executorplugins.ExecuteTemplateReply{
		Node: &wfv1.NodeResult{
			Phase:    status,
			Message:  job.Status.State.Message,
			Outputs:  nil,
			Progress: progress,
		},
		Requeue: requeue,
	})
}

func (ct *Controller) Response404(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotFound)

}

func InjectVcJobWithWorkflowName(job *batch.Job, workflowName string) {
	for _, task := range job.Spec.Tasks {
		if task.Template.ObjectMeta.Labels != nil {
			task.Template.ObjectMeta.Labels[LabelKeyWorkflow] = workflowName
		} else {
			task.Template.ObjectMeta.Labels = map[string]string{
				LabelKeyWorkflow: workflowName,
			}
		}
	}
}
