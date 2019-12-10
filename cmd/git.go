package cmd

import (
	"fmt"
	"os"
	"strings"

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

			fmt.Printf("%s\n", gitBranch())
		},
		TraverseChildren: true,
	}
	gitHashCmd = &cobra.Command{
		Use:   "hash",
		Short: "Show the current git hash",
		Long:  `Show the current git hash`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s\n", gitHash(gitOptions.gitHashLength))
		},
		TraverseChildren: true,
	}
)

func gitHash(length string) string {
	return execCmdOutput("git", "rev-parse", "--short="+length, "HEAD")
}

func gitBranch() string {
	if isGitHub() {
		tmp, exists := os.LookupEnv("GITHUB_REF")
		if exists && len(tmp) > 0 {
			return strings.TrimPrefix(tmp, "refs/heads/")
		}
	}
	if isGitLab() {
		tmp, exists := os.LookupEnv("CI_COMMIT_REF_NAME")
		if exists {
			return tmp
		}
	}
	return execCmdOutput("git", "rev-parse", "--abbrev-ref", "HEAD")
}

func isGitHub() bool {
	tmp, exists := os.LookupEnv("GITHUB_REF")
	return exists && len(tmp) > 0
}

func isGitLab() bool {
	tmp, exists := os.LookupEnv("GITLAB_CI")
	return exists && len(tmp) > 0
}
