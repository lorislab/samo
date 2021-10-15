package cmd

import (
	"github.com/lorislab/samo/tools"
	"github.com/spf13/cobra"
)

func createHealmReleaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "release",
		Short: "Release helm chart",
		Long:  `Download version of the helm chart and create final version`,
		Run: func(cmd *cobra.Command, args []string) {
			flags := helmFlags{}
			readOptions(&flags)
			project := loadProject(flags.Project)
			helmRelease(project, flags)
		},
		TraverseChildren: true,
	}

	return cmd
}

func helmRelease(pro *Project, flags helmFlags) {
	// clean helm dir
	healmClean(flags)
	// add custom helm repo
	healmAddRepo(flags)
	// update helm repo
	helmRepoUpdate()
	// download build version
	helmDownload(pro, flags)

	// update version to release version
	updateHelmChart(pro, flags)
	updateHelmValues(pro, flags)

	// package helm chart
	helmPackage(flags)
	// upload helm chart with release version
	helmPush(pro.ReleaseVersion(), pro, flags)
}

func helmDownload(project *Project, flags helmFlags) {
	var command []string
	command = append(command, "pull")
	command = append(command, flags.Repo+"/"+project.Name())
	command = append(command, "--version", project.Version())
	command = append(command, "--untar", "--untardir", flags.Dir)
	tools.ExecCmd("helm", command...)
}
