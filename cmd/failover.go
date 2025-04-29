package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// failoverCmd represents the failover command
var failoverCmd = &cobra.Command{
	Use:   "failover",
	Short: "Trigger redis failover",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("failover called")
	},
}

func init() {
	rootCmd.AddCommand(failoverCmd)
}
