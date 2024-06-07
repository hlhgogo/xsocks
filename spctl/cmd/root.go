package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "spctl",
	Short: "Use commend line managing simple.",
	Long:  `Use commend line managing simple.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
