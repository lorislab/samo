package cmd

import (
	"log"
	"os/exec"
	"strconv"

	"github.com/Masterminds/semver"
	"github.com/lorislab/samo/internal"
	"github.com/spf13/cobra"
)

func init() {
	createCmd.AddCommand(createReleaseCmd)
	addMavenFlags(createReleaseCmd)

	createCmd.AddCommand(createPatchCmd)
	addMavenFlags(createPatchCmd)
}

var (
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

			err := exec.Command("git", "tag", releaseVersion).Run()
			if err != nil {
				gitRollback(err)
			}

			newVersion := project.NextReleaseVersion()
			project.SetVersion(newVersion)

			err = exec.Command("git", "add", ".").Run()
			if err != nil {
				gitRollback(err)
			}
			err = exec.Command("git", "commit", "-m", "\"Create new release\"").Run()
			if err != nil {
				gitRollback(err)
			}
			err = exec.Command("git", "push", "origin", "refs/heads/*:refs/heads/*", "refs/tags/*:refs/tags/*").Run()
			if err != nil {
				gitRollback(err)
			}
		},
		TraverseChildren: true,
	}
	createPatchCmd = &cobra.Command{
		Use:   "patch",
		Short: "Create patch of the release",
		Long:  `Create patch of the release`,
		Run: func(cmd *cobra.Command, args []string) {
			project := internal.LoadMavenProject(mavenOptions.filename)
			ver, e := semver.NewVersion(project.Version())
			if e != nil {
				panic(e)
			}

			branchName := strconv.FormatInt(ver.Major(), 10) + "." + strconv.FormatInt(ver.Minor(), 10)
			err := exec.Command("git", "checkout", "-b", branchName, project.Version()).Run()
			if err != nil {
				gitRollback(err)
			}

			err = exec.Command("git", "add", ".").Run()
			if err != nil {
				gitRollback(err)
			}
			err = exec.Command("git", "commit", "-m", "\"Create new release\"").Run()
			if err != nil {
				gitRollback(err)
			}
			err = exec.Command("git", "push", "origin", "refs/heads/*:refs/heads/*", "refs/tags/*:refs/tags/*").Run()
			if err != nil {
				gitRollback(err)
			}
		},
		TraverseChildren: true,
	}
)

func gitRollback(e error) {
	log.Fatal(e)
	err := exec.Command("rm", "-f", "o.git/index.lock").Run()
	if err != nil {
		panic(err)
	}
	panic(e)
}
