package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/lorislab/samo/internal"

	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

func init() {
	gitCmd.AddCommand(gitBranchCmd)

	gitCmd.AddCommand(gitBuildVersionCmd)
	addFlag(gitBuildVersionCmd, "build-number-prefix", "", "rc", "the build number prefix")
	addIntFlag(gitBuildVersionCmd, "build-number-length", "", 3, "the build number length")
	hashLength := addGitHashLength(gitBuildVersionCmd, "hash-length", "")

	gitCmd.AddCommand(gitVersionCmd)
	addFlagRef(gitVersionCmd, hashLength)

	gitCmd.AddCommand(gitHashCmd)
	addFlagRef(gitHashCmd, hashLength)

	gitCmd.AddCommand(gitCreateReleaseCmd)
	addBoolFlag(gitCreateReleaseCmd, "release-major", "a", false, "create a major release")
	addFlag(gitCreateReleaseCmd, "release-tag", "t", "", "the tag release version")
	addFlag(gitCreateReleaseCmd, "release-tag-message", "m", "", "the release tag message")
	addFlagRef(gitCreateReleaseCmd, hashLength)
	addBoolFlag(gitCreateReleaseCmd, "release-skip-push", "", false, "skip git push release")

	gitCmd.AddCommand(gitReleaseVersionCmd)
	addFlagRef(gitReleaseVersionCmd, hashLength)

	gitCmd.AddCommand(gitCreatePatchCmd)
	addFlagRef(gitCreatePatchCmd, hashLength)
	addFlagRequired(gitCreatePatchCmd, "patch-tag", "t", "", "the tag version for the patch branch")
	addFlag(gitCreatePatchCmd, "patch-branch-prefix", "", "", "patch branch prefix")
	addBoolFlag(gitCreatePatchCmd, "patch-skip-push", "", false, "skip git push patch branch")

}

type gitFlags struct {
	HashLength        int    `mapstructure:"hash-length"`
	ReleaseTag        string `mapstructure:"release-tag"`
	PatchTag          string `mapstructure:"patch-tag"`
	ReleaseMajor      bool   `mapstructure:"release-major"`
	BuildNumberPrefix string `mapstructure:"build-number-prefix"`
	BuildNumberLength int    `mapstructure:"build-number-length"`
	ReleaseTagMessage string `mapstructure:"release-tag-message"`
	PatchBranchPrefix string `mapstructure:"patch-branch-prefix"`
	PatchSkipPush     bool   `mapstructure:"patch-skip-push"`
	ReleaseSkipPush   bool   `mapstructure:"release-skip-push"`
}

var (
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
			_, project := readGitOptions()
			fmt.Printf("%s\n", project.Hash())
		},
		TraverseChildren: true,
	}
	gitVersionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show the current git version",
		Long:  `Show the current git version`,
		Run: func(cmd *cobra.Command, args []string) {
			_, project := readGitOptions()
			projectVersion(project)
		},
		TraverseChildren: true,
	}
	gitBuildVersionCmd = &cobra.Command{
		Use:   "build-version",
		Short: "Show the current git build version",
		Long:  `Show the current git build version`,
		Run: func(cmd *cobra.Command, args []string) {
			options, project := readGitOptions()
			projectBuildVersion(project, options.HashLength, options.BuildNumberLength, options.BuildNumberPrefix)
		},
		TraverseChildren: true,
	}
	gitCreateReleaseCmd = &cobra.Command{
		Use:   "create-release",
		Short: "Create release of the current project and state",
		Long:  `Create release of the current project and state`,
		Run: func(cmd *cobra.Command, args []string) {
			options, project := readGitOptions()
			projectCreateRelease(project, options.ReleaseTagMessage, "", options.ReleaseMajor, options.ReleaseSkipPush)
		},
		TraverseChildren: true,
	}
	gitReleaseVersionCmd = &cobra.Command{
		Use:   "release-version",
		Short: "Show release of the current project and state",
		Long:  `Show release of the current project and state`,
		Run: func(cmd *cobra.Command, args []string) {
			_, project := readGitOptions()
			projectReleaseVersion(project)
		},
		TraverseChildren: true,
	}
	gitCreatePatchCmd = &cobra.Command{
		Use:   "create-patch",
		Short: "Create patch of the release",
		Long:  `Create patch of the release`,
		Run: func(cmd *cobra.Command, args []string) {
			options, project := readGitOptions()
			projectCreatePatch(project, "", options.PatchTag, options.PatchBranchPrefix, options.PatchSkipPush)
		},
		TraverseChildren: true,
	}
)

func readGitOptions() (gitFlags, *internal.GitProject) {
	options := gitFlags{}
	err := viper.Unmarshal(&options)
	if err != nil {
		panic(err)
	}
	return options, internal.LoadGitProject(options.HashLength)
}

func gitBranch() string {
	if isGitHub() {
		tmp, exists := os.LookupEnv("GITHUB_REF")
		if exists && len(tmp) > 0 {
			return strings.TrimPrefix(tmp, "refs/heads/")
		}
	}
	if isGitLab() {
		tmp, exists := os.LookupEnv("CI_COMMIT_REF_SLUG")
		if exists {
			return tmp
		}
	}
	return internal.ExecCmdOutput("git", "rev-parse", "--abbrev-ref", "HEAD")
}
