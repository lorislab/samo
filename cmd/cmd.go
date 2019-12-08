package cmd

import "github.com/spf13/cobra"

// Main the commands method
func Main(rootCmd *cobra.Command) {
	rootCmd.AddCommand(mvnCmd)
	rootCmd.AddCommand(gitCmd)
	rootCmd.AddCommand(dockerCmd)
	rootCmd.AddCommand(createCmd)
}
