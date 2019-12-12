package cmd

import (
	"fmt"
	"strconv"

	"github.com/Masterminds/semver"
	"github.com/lorislab/samo/internal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type mavenFlags struct {
	filename      string
	patchMsg      string
	devMsg        string
	tag           string
	major         bool
	gitHashLength string
}

func init() {
	mvnCmd.AddCommand(mvnVersionCmd)
	addMavenFlags(mvnVersionCmd)
	mvnCmd.AddCommand(mvnSetSnapshotCmd)
	addMavenFlags(mvnSetSnapshotCmd)
	mvnCmd.AddCommand(mvnSetReleaseCmd)
	addMavenFlags(mvnSetReleaseCmd)
	mvnCmd.AddCommand(mvnSetHashCmd)
	addMavenFlags(mvnSetHashCmd)
	mvnSetHashCmd.Flags().StringVarP(&(mavenOptions.gitHashLength), "length", "l", "7", "The git hash length")

	mvnCmd.AddCommand(mvnCreateReleaseCmd)
	addMavenFlags(mvnCreateReleaseCmd)
	mvnCreateReleaseCmd.Flags().StringVarP(&(mavenOptions.devMsg), "message", "m", "Create new development version", "Commit message for new development version")
	mvnCreateReleaseCmd.Flags().BoolVarP(&(mavenOptions.major), "major", "a", false, "Create a major release")

	mvnCmd.AddCommand(mvnCreatePatchCmd)
	addMavenFlags(mvnCreatePatchCmd)
	mvnCreatePatchCmd.Flags().StringVarP(&(mavenOptions.tag), "tag", "t", "", "The tag version for the patch branch")
	mvnCreatePatchCmd.Flags().StringVarP(&(mavenOptions.patchMsg), "message", "m", "Create new patch version", "Commit message for new patch version")
	err := mvnCreatePatchCmd.MarkFlagRequired("tag")
	if err != nil {
		log.Panic(err)
	}
}

func addMavenFlags(command *cobra.Command) {
	command.Flags().StringVarP(&(mavenOptions.filename), "file", "f", "pom.xml", "The maven project file")
}

var (
	mavenOptions = mavenFlags{}

	// MvnCmd the maven command
	mvnCmd = &cobra.Command{
		Use:              "maven",
		Short:            "Maven operation",
		Long:             `Tasks for the maven project`,
		TraverseChildren: true,
	}
	mvnVersionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show the maven project version",
		Long:  `Tasks to show the maven project version`,
		Run: func(cmd *cobra.Command, args []string) {
			project := internal.LoadMavenProject(mavenOptions.filename)
			fmt.Printf("%s\n", project.Version())
		},
		TraverseChildren: true,
	}
	mvnSetSnapshotCmd = &cobra.Command{
		Use:   "set-snapshot",
		Short: "Set the maven project version to snapshot",
		Long:  `Set the maven project version to snapshot`,
		Run: func(cmd *cobra.Command, args []string) {
			project := internal.LoadMavenProject(mavenOptions.filename)
			version := project.SetPrerelease("SNAPSHOT")
			project.SetVersion(version)
		},
		TraverseChildren: true,
	}
	mvnSetReleaseCmd = &cobra.Command{
		Use:   "set-release",
		Short: "Set the maven project version to release",
		Long:  `Set the maven project version to release`,
		Run: func(cmd *cobra.Command, args []string) {
			project := internal.LoadMavenProject(mavenOptions.filename)
			version := project.ReleaseVersion()
			project.SetVersion(version)
		},
		TraverseChildren: true,
	}
	mvnSetHashCmd = &cobra.Command{
		Use:   "set-hash",
		Short: "Set the maven project version to git hash",
		Long:  `Set the maven project version to git hash`,
		Run: func(cmd *cobra.Command, args []string) {
			project := internal.LoadMavenProject(mavenOptions.filename)
			hash := gitHash(mavenOptions.gitHashLength)
			version := project.SetPrerelease(hash)
			project.SetVersion(version)
		},
		TraverseChildren: true,
	}
	mvnCreateReleaseCmd = &cobra.Command{
		Use:   "create-release",
		Short: "Create release of the current project and state",
		Long:  `Create release of the current project and state`,
		Run: func(cmd *cobra.Command, args []string) {
			project := internal.LoadMavenProject(mavenOptions.filename)
			releaseVersion := project.ReleaseVersion()

			execGitCmd("git", "tag", releaseVersion)

			newVersion := project.NextReleaseVersion(mavenOptions.major)
			project.SetVersion(newVersion)

			execGitCmd("git", "add", ".")
			execGitCmd("git", "commit", "-m", mavenOptions.devMsg+" ["+newVersion+"]")
			execGitCmd("git", "push", "origin", "refs/heads/*:refs/heads/*", "refs/tags/*:refs/tags/*")
		},
		TraverseChildren: true,
	}
	mvnCreatePatchCmd = &cobra.Command{
		Use:   "create-patch",
		Short: "Create patch of the release",
		Long:  `Create patch of the release`,
		Run: func(cmd *cobra.Command, args []string) {

			tagVer, e := semver.NewVersion(mavenOptions.tag)
			if e != nil {
				log.Panic(e)
			}

			branchName := strconv.FormatInt(tagVer.Major(), 10) + "." + strconv.FormatInt(tagVer.Minor(), 10)
			execGitCmd("git", "checkout", "-b", branchName, mavenOptions.tag)

			project := internal.LoadMavenProject(mavenOptions.filename)
			patch := project.NextPatchVersion()
			project.SetVersion(patch)

			execGitCmd("git", "add", ".")
			execGitCmd("git", "commit", "-m", mavenOptions.patchMsg+" ["+patch+"]")
			execGitCmd("git", "push", "origin", "refs/heads/*:refs/heads/*")

		},
		TraverseChildren: true,
	}
)
