package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/viper"

	"github.com/Masterminds/semver"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var numberRegex *regexp.Regexp

func init() {
	numberRegex = regexp.MustCompile("[0-9]+")

	gitCmd.AddCommand(gitBranchCmd)
	gitCmd.AddCommand(gitBuildVersionCmd)
	addFlag(gitBuildVersionCmd, "build-number-prefix", "", "rc", "the build number prefix")
	addIntFlag(gitBuildVersionCmd, "build-number-length", "", 3, "the build number length")
	hashLength := addGitHashLength(gitBuildVersionCmd, "hash-length", "l")

	gitCmd.AddCommand(gitHashCmd)
	addFlagRef(gitHashCmd, hashLength)

	gitCmd.AddCommand(gitCreateReleaseCmd)
	addBoolFlag(gitCreateReleaseCmd, "release-major", "a", false, "create a major release")
	addFlag(gitCreateReleaseCmd, "release-tag", "t", "", "the tag release version")
	addFlag(gitCreateReleaseCmd, "release-tag-message", "m", "", "the release tag message")
	addFlagRef(gitCreateReleaseCmd, hashLength)

	gitCmd.AddCommand(gitReleaseVersionCmd)
	addFlagRef(gitReleaseVersionCmd, hashLength)

	gitCmd.AddCommand(gitCreatePatchCmd)
	addFlagRequired(gitCreatePatchCmd, "patch-tag", "t", "", "the tag version for the patch branch")
}

type gitFlags struct {
	HashLength        int    `mapstructure:"hash-length"`
	ReleaseTag        string `mapstructure:"release-tag"`
	PatchTag          string `mapstructure:"patch-tag"`
	Major             bool   `mapstructure:"release-major"`
	BuildNumberPrefix string `mapstructure:"build-number-prefix"`
	BuildNumberLength int    `mapstructure:"build-number-length"`
	ReleaseTagMessage string `mapstructure:"release-tag-message"`
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
	gitBuildVersionCmd = &cobra.Command{
		Use:   "build-version",
		Short: "Show the current git build version",
		Long:  `Show the current git build version`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readGitOptions()
			lastTag, count, hash := gitCommit(options.HashLength)
			ver := createBuildVersion(lastTag, count, hash, options.BuildNumberPrefix, options.BuildNumberLength)
			fmt.Printf("%s\n", ver.String())
		},
		TraverseChildren: true,
	}
	gitCreateReleaseCmd = &cobra.Command{
		Use:   "create-release",
		Short: "Create release of the current project and state",
		Long:  `Create release of the current project and state`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readGitOptions()
			ver := gitReleaseVersion(options.ReleaseTag, options.HashLength, options.Major)
			msg := options.ReleaseTagMessage
			if len(msg) == 0 {
				msg = ver
			}
			execGitCmd("git", "tag", "-a", ver, "-m", msg)
			execGitCmd("git", "push", "--tag")
			log.Infof("New release [%s] created.", ver)
		},
		TraverseChildren: true,
	}
	gitReleaseVersionCmd = &cobra.Command{
		Use:   "release-version",
		Short: "Show release of the current project and state",
		Long:  `Show release of the current project and state`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readGitOptions()
			ver := gitReleaseVersion(options.ReleaseTag, options.HashLength, options.Major)
			fmt.Printf("%s\n", ver)
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
			branchName := createPatchBranchName(tagVer)
			execGitCmd("git", "checkout", "-b", branchName, options.PatchTag)
			execGitCmd("git", "push", "origin", "refs/heads/*:refs/heads/*")
			log.Infof("New patch branch for version [%s] created.", branchName)

		},
		TraverseChildren: true,
	}
)

// <VERSION>(+1)-<BUILD>-<HASH>
func gitReleaseVersion(tag string, hashLength int, major bool) string {
	ver := tag
	if len(ver) == 0 {
		lastTag, _, _ := gitCommit(hashLength)
		v := nextReleaseVersion(createVersion(lastTag), major)
		ver = v.String()
	}
	return ver
}

func readGitOptions() gitFlags {
	gitOptions := gitFlags{}
	err := viper.Unmarshal(&gitOptions)
	if err != nil {
		panic(err)
	}
	return gitOptions
}

func gitCommit(length int) (string, string, string) {
	// search for latest annotated tag
	lastTag, err := execCmdOutputErr("git", "describe", "--abbrev=0")
	log.Debugf("Last tag %s", lastTag)
	if err == nil {
		// get metadata from the git describe
		describe, err := execCmdOutputErr("git", "describe", "--long", "--abbrev="+strconv.Itoa(length))
		if err == nil {
			describe = strings.TrimPrefix(describe, lastTag+"-")
			items := strings.Split(describe, "-")
			return lastTag, items[0], items[1]
		}
	}
	// not tag found in the git repository
	lastTag = "0.0.0"
	count := "0"
	// git commit hash
	hash, err := execCmdOutputErr("git", "rev-parse", "--short="+strconv.Itoa(length), "HEAD")
	if err != nil {
		hash = lpad("", "0", length)
	} else {
		// git commit count in the branch
		tmp, err := execCmdOutputErr("git", "rev-list", "HEAD", "--count")
		if err == nil {
			count = tmp
		}
	}
	return lastTag, count, hash
}

func gitHash(length int) string {
	if length > 0 {
		return execCmdOutput("git", "rev-parse", "--short="+strconv.Itoa(length), "HEAD")
	}
	return execCmdOutput("git", "rev-parse", "HEAD")
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
