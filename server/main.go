package main

import (
	"argo-volcano-executor-plugin/server/options"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
)

const (
	// CLIName is the name of the CLI : argo volcano plugin
	CLIName = "avp"
)

var logFlushFreq = pflag.Duration("log-flush-frequency", 5*time.Second, "Maximum number of seconds between log flushes")

func main() {
	// flag.InitFlags()
	klog.InitFlags(nil)

	// The default klog flush interval is 30 seconds, which is frighteningly long.
	go wait.Until(klog.Flush, *logFlushFreq, wait.NeverStop)
	defer klog.Flush()

	rootCmd := cobra.Command{
		Use: "avp",
	}

	config := options.NewConfig()
	config.AddFlags(pflag.CommandLine)

	rootCmd.AddCommand(runServer(config))

	rootCmd.AddCommand(NewVersionCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Failed to execute command: %v\n", err)
		os.Exit(2)
	}
}
