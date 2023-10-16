package main

import (
	"argo-volcano-executor-plugin/controller"
	"argo-volcano-executor-plugin/pkg/kube"
	"argo-volcano-executor-plugin/server/options"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"net/http"
)

func runServer(config *options.Config) *cobra.Command {
	rootCmd := cobra.Command{
		Use:   "server",
		Short: "argo volcano job plugin",
		Long:  `a argo step that can run a volcano job`,
		RunE: func(cmd *cobra.Command, args []string) error {

			pflag.Parse()
			fmt.Println("### Listen on: ", config.Port)
			return runPlugin(config)
		},
	}
	return &rootCmd
}

func runK8sPlugin(config *options.Config) error {
	//todo
	if config.PluginName != options.K8sPluginName {
		return errors.New("Not valid plugin name. must be k8s here.")
	}
	restConfig, err := kube.BuildConfig(config.KubeClientOptions)
	if err != nil {
		return fmt.Errorf("unable to build k8s config: %v", err)
	}
	ct := &controller.K8sController{}
	kubeClient := getKubeClient(restConfig)
	ct.KubeClient = kubeClient

	return nil
}

func runPlugin(config *options.Config) error {

	if config.PluginName == options.K8sPluginName {
		return runK8sPlugin(config)
	}
	restConfig, err := kube.BuildConfig(config.KubeClientOptions)
	if err != nil {
		return fmt.Errorf("unable to build k8s config: %v", err)
	}
	ct := &controller.VCController{}

	vcClient := getVolcanoClient(restConfig)
	kubeClient := getKubeClient(restConfig)
	argoClient := getArgoClient(restConfig)

	ct.VcClient = vcClient
	ct.KubeClient = kubeClient
	ct.ArgoClient = argoClient

	router := gin.Default()

	// Query string parameters are parsed using the existing underlying request object.
	// The request responds to a url matching:  /welcome?firstname=Jane&lastname=Doe
	router.GET("/welcome", func(c *gin.Context) {
		firstname := c.DefaultQuery("firstname", "Guest")
		lastname := c.Query("lastname") // shortcut for c.Request.URL.Query().Get("lastname")

		c.String(http.StatusOK, "Hello %s %s", firstname, lastname)
	})
	router.POST("/api/v1/template.execute", ct.ExecuteVolcanoJob)
	return router.Run(fmt.Sprintf(":%d", config.Port))
}
