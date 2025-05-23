package cmd

import (
	"os"

	"github.com/seeker89/redis-resiliency-toolkit/pkg/config"
	"github.com/seeker89/redis-resiliency-toolkit/pkg/printer"
	"github.com/spf13/cobra"
)

var Version, Build string
var cfg config.RRConfig
var prtr *printer.Printer
var CMD_PREFIX = "RR_"

var rootCmd = &cobra.Command{
	Use:   "rr",
	Short: "Verify resiliency of your redis setup",
	Long:  ``,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		prtr = printer.NewPrinter(cfg.Format, cfg.Pretty, os.Stdout)
	},
	// this is a rather annoying part of Cobra
	// when using RunE, on any error, it will print the help output
	// which is what you want when arguments validation fails, but not on actual runtim error
	// so this disables
	SilenceUsage: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version, build string) {
	Version = version
	Build = build
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// format options
	rootCmd.PersistentFlags().BoolVarP(&cfg.Verbose, "verbose", "v", false, "Make the output verbose")
	rootCmd.PersistentFlags().BoolVarP(&cfg.Pretty, "pretty", "p", false, "Make the output pretty")
	rootCmd.PersistentFlags().StringVarP(&cfg.Format, "output", "o", "json", "Output format (json, text, wide)")
	// kubernetes options
	rootCmd.PersistentFlags().StringVar(&cfg.Kubeconfig, "kubeconfig", os.Getenv("KUBECONFIG"), "Path to a kubeconfig file. Leave empty for in-cluster. (KUBECONFIG)")
	rootCmd.PersistentFlags().StringVar(&cfg.Namespace, "namespace", os.Getenv("NAMESPACE"), "Limit Kubernetes actions to only this namespace (NAMESPACE)")
}
