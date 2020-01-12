package cmd

import (
	"fmt"
	"github.com/spf13/viper"
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
	addFlag(gitHashCmd, "hash-length", "l", "7", "The git hash length")

	gitCmd.AddCommand(gitCreateReleaseCmd)
	addBoolFlag(gitCreateReleaseCmd, "release-major", "a", false, "Create a major release")
	addFlag(gitCreateReleaseCmd, "release-tag", "t", "", "The tag release version")

	gitCmd.AddCommand(gitCreatePatchCmd)
	addFlagRequired(gitCreatePatchCmd, "patch-tag", "t", "", "The tag version for the patch branch")
}

type gitFlags struct {
	HashLength string `mapstructure:"hash-length"`
	ReleaseTag string `mapstructure:"release-tag"`
	PatchTag   string `mapstructure:"patch-tag"`
	Major      bool   `mapstructure:"release-major"`
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
			options := readGitOptions()
			fmt.Printf("%s\n", gitHash(options.HashLength))
		},
		TraverseChildren: true,
	}
	gitCreateReleaseCmd = &cobra.Command{
		Use:   "create-release",
		Short: "Create release of the current project and state",
		Long:  `Create release of the current project and state`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readGitOptions()
			ver := options.ReleaseTag
			if len(ver) == 0 {
				lastTag := execCmdOutput("git", "describe", "--tags")
				log.Infof("Last release version [%s]", lastTag)
				tagVer, e := semver.NewVersion(lastTag)
				if e != nil {
					log.Panic(e)
				}
				if options.Major {
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
		Use:   "create-patch",
		Short: "Create patch of the release",
		Long:  `Create patch of the release`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readGitOptions()
			tagVer, e := semver.NewVersion(options.PatchTag)
			if e != nil {
				log.Panic(e)
			}
			if tagVer.Patch() != 0 {
				log.Errorf("Can not created patch branch from the patch version  [%s]!", tagVer.Original())
				os.Exit(0)
			}
			branchName := strconv.FormatInt(tagVer.Major(), 10) + "." + strconv.FormatInt(tagVer.Minor(), 10)
			execGitCmd("git", "checkout", "-b", branchName, options.PatchTag)
			execGitCmd("git", "push", "origin", "refs/heads/*:refs/heads/*")
			log.Infof("New patch branch for version [%s] created.", branchName)

		},
		TraverseChildren: true,
	}
)

func readGitOptions() gitFlags {
	gitOptions := gitFlags{}
	err := viper.Unmarshal(&gitOptions)
	if err != nil {
		panic(err)
	}
	return gitOptions
}

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

func execGitCmd(name string, arg ...string) {
	err := execCmdErr(name, arg...)
	if err != nil {
		execCmd("rm", "-f", "o.git/index.lock")
	}
}
