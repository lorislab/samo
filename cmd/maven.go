package cmd

import (
	"fmt"

	"github.com/lorislab/samo/internal"
	"github.com/spf13/cobra"
)

type mavenFlags struct {
	filename string
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
	addGitFlags(mvnSetHashCmd)
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
			hash := internal.GitHash(gitOptions.gitHashLength)
			fmt.Printf("%s\n", hash)
			version := project.SetPrerelease(hash)
			project.SetVersion(version)
		},
		TraverseChildren: true,
	}
)
