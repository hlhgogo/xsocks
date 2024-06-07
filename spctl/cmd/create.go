package cmd

import (
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create resource",
	Long:  `Create resource`,
}

func init() {
	rootCmd.AddCommand(createCmd)
}
