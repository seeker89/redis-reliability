package cmd

import (
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Run: func(cmd *cobra.Command, args []string) {
		prtr.Itemise = true
		prtr.Print([]map[string]string{
			{
				"version": Version,
				"build":   Build,
			},
		}, []string{})
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
