package cmd

import (
	"fmt"

	"github.com/lorislab/samo/internal"
	"github.com/spf13/cobra"
)

func init() {
	gitCmd.AddCommand(gitBranchCmd)
	gitCmd.AddCommand(gitHashCmd)
	addGitFlags(gitHashCmd)
}

func addGitFlags(command *cobra.Command) {
	command.Flags().StringVarP(&(gitOptions.gitHashLength), "length", "l", "7", "The git hash length")
}

type gitFlags struct {
	gitHashLength string
}

var (
	gitOptions = gitFlags{}

	gitCmd = &cobra.Command{
		Use:              "git",
		Short:            "Git operation",
		Long:             `Git operation`,
		TraverseChildren: true,
	}
	gitBranchCmd = &cobra.Command{
		Use:   "branch",
		Short: "Show the current git branch",
		Long:  `Show the current git branch`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s\n", internal.GitBranch())
		},
		TraverseChildren: true,
	}
	gitHashCmd = &cobra.Command{
		Use:   "hash",
		Short: "Show the current git hash",
		Long:  `Show the current git hash`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s\n", internal.GitHash(gitOptions.gitHashLength))
		},
		TraverseChildren: true,
	}
)
