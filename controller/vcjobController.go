package controller

import (
	"fmt"
	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	executorplugins "github.com/argoproj/argo-workflows/v3/pkg/plugins/executor"
	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"net/http"
	"time"
	"volcano.sh/apis/pkg/client/clientset/versioned"
)

type Controller struct {
	VcClient   *versioned.Clientset
	KubeClient *kubernetes.Clientset
}

func (ct *Controller) ExecuteVolcanoJob(ctx *gin.Context) {
	c := &executorplugins.ExecuteTemplateArgs{}
	err := ctx.BindJSON(&c)
	if err != nil {
		klog.Error(err)
		return
	}
	ob, _ := c.Template.Plugin.MarshalJSON()
	fmt.Println(ob)

}

func (ct *Controller) Response404(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotFound)

}

func (ct *Controller) Response(ctx *gin.Context) {
	progress , _:=wfv1.NewProgress(0, 5)

	ctx.JSON(http.StatusOK, &executorplugins.ExecuteTemplateResponse{
		Body: executorplugins.ExecuteTemplateReply{
			Node: &wfv1.NodeResult{
				Phase: wfv1.NodePending,
				Message: "",
				Outputs: nil,
				Progress: progress,
			},
			Requeue: &metav1.Duration{
				Duration: 10 * time.Second,
			},
		},
	})
}
