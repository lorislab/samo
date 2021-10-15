package cmd

import (
	"github.com/lorislab/samo/tools"
	"github.com/spf13/cobra"
)

type helmReleaseFlags struct {
	Helm                 helmFlags `mapstructure:",squash"`
	ChartReleaseTemplate string    `mapstructure:"chart-release-template-list"`
}

func createHealmReleaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "release",
		Short: "Release helm chart",
		Long:  `Download version of the helm chart and create final version`,
		Run: func(cmd *cobra.Command, args []string) {
			flags := helmReleaseFlags{}
			readOptions(&flags)
			project := loadProject(flags.Helm.Project)
			helmRelease(project, flags)
		},
		TraverseChildren: true,
	}

	addStringFlag(cmd, "chart-release-template-list", "", "version={{ .Release }},appVersion={{ .Release }},name={{ .Name }}", `list of key value to be replaced in the Chart.yaml
	Values: Name,Hash,Branch,Tag,Count,Version,Release. 
	Example: version={{ .Release }},appVersion={{ .Release }}`)

	return cmd
}

func helmRelease(pro *Project, flags helmReleaseFlags) {
	// clean helm dir
	healmClean(flags.Helm)

	// add custom helm repo
	healmAddRepo(flags.Helm)

	// update helm repo
	helmRepoUpdate()

	// download build version
	helmDownload(pro, flags.Helm)

	// update version to release version
	updateHelmChart(pro, flags.Helm, flags.ChartReleaseTemplate)
	updateHelmValues(pro, flags.Helm)

	// package helm chart
	helmPackage(pro, flags.Helm)

	// upload helm chart with release version
	helmPush(pro.Release(), pro, flags.Helm)
}

func helmDownload(project *Project, flags helmFlags) {
	var command []string
	command = append(command, "pull")
	command = append(command, flags.Repo+"/"+project.Name())
	command = append(command, "--version", project.Version())
	command = append(command, "--untar", "--untardir", flags.Dir)
	tools.ExecCmd("helm", command...)
}
