package cmd

import (
	"strconv"

	"github.com/Masterminds/semver"
	"github.com/lorislab/samo/internal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	createCmd.AddCommand(createReleaseCmd)
	addMavenFlags(createReleaseCmd)
	createReleaseCmd.Flags().StringVarP(&(createOptions.devMsg), "message", "m", "Create new development version", "Commit message for new development version")
	createReleaseCmd.Flags().BoolVarP(&(createOptions.major), "major", "a", false, "Create a major release")

	createCmd.AddCommand(createPatchCmd)
	addMavenFlags(createPatchCmd)
	createPatchCmd.Flags().StringVarP(&(createOptions.tag), "tag", "t", "", "The tag versoin for the patch branch")
	createPatchCmd.Flags().StringVarP(&(createOptions.patchMsg), "message", "m", "Create new patch version", "Commit message for new patch version")
	createPatchCmd.MarkFlagRequired("tag")
}

type createFlags struct {
	patchMsg string
	devMsg   string
	tag      string
	major    bool
}

var (
	createOptions = createFlags{}

	createCmd = &cobra.Command{
		Use:              "create",
		Short:            "Create operation",
		Long:             `Create operation`,
		TraverseChildren: true,
	}
	createReleaseCmd = &cobra.Command{
		Use:   "release",
		Short: "Create release of the current project and state",
		Long:  `Create release of the current project and state`,
		Run: func(cmd *cobra.Command, args []string) {
			project := internal.LoadMavenProject(mavenOptions.filename)
			releaseVersion := project.ReleaseVersion()

			execGitCmd("git", "tag", releaseVersion)

			newVersion := project.NextReleaseVersion(createOptions.major)
			project.SetVersion(newVersion)

			execGitCmd("git", "add", ".")
			execGitCmd("git", "commit", "-m", "\""+createOptions.devMsg+" ["+newVersion+"]\"")
			execGitCmd("git", "push", "origin", "refs/heads/*:refs/heads/*", "refs/tags/*:refs/tags/*")
		},
		TraverseChildren: true,
	}
	createPatchCmd = &cobra.Command{
		Use:   "patch",
		Short: "Create patch of the release",
		Long:  `Create patch of the release`,
		Run: func(cmd *cobra.Command, args []string) {

			tagVer, e := semver.NewVersion(createOptions.tag)
			if e != nil {
				log.Panic(e)
			}

			branchName := strconv.FormatInt(tagVer.Major(), 10) + "." + strconv.FormatInt(tagVer.Minor(), 10)
			execGitCmd("git", "checkout", "-b", branchName, createOptions.tag)

			project := internal.LoadMavenProject(mavenOptions.filename)
			patch := project.NextPatchVersion()
			project.SetVersion(patch)

			execGitCmd("git", "add", ".")
			execGitCmd("git", "commit", "-m", "\""+createOptions.patchMsg+" ["+patch+"]\"")
			execGitCmd("git", "push", "origin", "refs/heads/*:refs/heads/*")

		},
		TraverseChildren: true,
	}
)

func execGitCmd(name string, arg ...string) {
	err := execCmdErr(name, arg...)
	if err != nil {
		execCmd("rm", "-f", "o.git/index.lock")
	}
}
