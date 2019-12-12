package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Masterminds/semver"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	gitCmd.AddCommand(gitBranchCmd)
	gitCmd.AddCommand(gitHashCmd)
	addGitFlags(gitHashCmd)

	gitCmd.AddCommand(gitCreateReleaseCmd)
	gitCreateReleaseCmd.Flags().StringVarP(&(gitOptions.devMsg), "message", "m", "Create new development version", "Commit message for new development version")
	gitCreateReleaseCmd.Flags().BoolVarP(&(gitOptions.major), "major", "a", false, "Create a major release")
	gitCreateReleaseCmd.Flags().StringVarP(&(gitOptions.tag), "tag", "t", "", "The tag release version")

	gitCmd.AddCommand(gitCreatePatchCmd)
	gitCreatePatchCmd.Flags().StringVarP(&(gitOptions.tag), "tag", "t", "", "The tag version for the patch branch")
	gitCreatePatchCmd.Flags().StringVarP(&(gitOptions.patchMsg), "message", "m", "Create new patch version", "Commit message for new patch version")
	gitCreatePatchCmd.MarkFlagRequired("tag")
}

func addGitFlags(command *cobra.Command) {
	command.Flags().StringVarP(&(gitOptions.gitHashLength), "length", "l", "7", "The git hash length")
}

type gitFlags struct {
	gitHashLength string
	patchMsg      string
	devMsg        string
	tag           string
	major         bool
	branch        string
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
	gitCreateReleaseCmd = &cobra.Command{
		Use:   "create-release",
		Short: "Create release of the current project and state",
		Long:  `Create release of the current project and state`,
		Run: func(cmd *cobra.Command, args []string) {
			ver := gitOptions.tag
			if len(ver) == 0 {
				lastTag := execCmdOutput("git", "describe", "--tags")
				log.Infof("Last release version [%s]", lastTag)
				tagVer, e := semver.NewVersion(lastTag)
				if e != nil {
					log.Panic(e)
				}
				if gitOptions.major {
					if tagVer.Patch() != 0 {
						log.Errorf("Can not created major release from the patch version  [%s]!", lastTag)
						os.Exit(0)
					}
					tmp := tagVer.IncMajor()
					ver = tmp.Original()
				} else {
					if tagVer.Patch() == 0 {
						tmp := tagVer.IncMinor()
						ver = tmp.Original()
					} else {
						tmp := tagVer.IncPatch()
						ver = tmp.Original()
					}
				}

			}
			execGitCmd("git", "tag", ver)
			execGitCmd("git", "push", "--tag")
			log.Infof("New release [%s] created.", ver)
		},
		TraverseChildren: true,
	}
	gitCreatePatchCmd = &cobra.Command{
		Use:   "craete-patch",
		Short: "Create patch of the release",
		Long:  `Create patch of the release`,
		Run: func(cmd *cobra.Command, args []string) {
			tagVer, e := semver.NewVersion(gitOptions.tag)
			if e != nil {
				log.Panic(e)
			}
			if tagVer.Patch() != 0 {
				log.Errorf("Can not created patch branch from the patch version  [%s]!", tagVer.Original())
				os.Exit(0)
			}
			branchName := strconv.FormatInt(tagVer.Major(), 10) + "." + strconv.FormatInt(tagVer.Minor(), 10)
			execGitCmd("git", "checkout", "-b", branchName, mavenOptions.tag)
			execGitCmd("git", "push", "origin", "refs/heads/*:refs/heads/*")
			log.Info("New patch branch for version [%s] created.", branchName)

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

func execGitCmd(name string, arg ...string) {
	err := execCmdErr(name, arg...)
	if err != nil {
		execCmd("rm", "-f", "o.git/index.lock")
	}
}
