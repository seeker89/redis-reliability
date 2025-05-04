package cmd

import (
	"os"

	"github.com/seeker89/redis-resiliency-toolkit/pkg/config"
	"github.com/spf13/cobra"
)

var cfg config.RRTConfig

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "redis-resiliency-toolkit",
	Short: "Test the resiliency of your redis setup",
	Long:  ``,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&cfg.Verbose, "verbose", "v", false, "Make the output verbose")
	rootCmd.PersistentFlags().StringVar(&cfg.Kubeconfig, "kube-config", "", "Path to a kubeconfig file. Leave empty for in-cluster")
}
