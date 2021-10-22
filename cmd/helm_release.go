package cmd

import (
	"github.com/lorislab/samo/log"
	"github.com/lorislab/samo/tools"
	"github.com/spf13/cobra"
)

type helmReleaseFlags struct {
	Helm                  helmFlags `mapstructure:",squash"`
	ChartReleaseTemplate  string    `mapstructure:"helm-chart-release-template-list"`
	ValuesReleaseTemplate string    `mapstructure:"helm-values-release-template-list"`
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

	addStringFlag(cmd, "helm-chart-release-template-list", "", "version={{ .Release }},appVersion={{ .Release }},name={{ .Name }}", `list of key value to be replaced in the Chart.yaml
	Values: `+templateValues+`
	Example: version={{ .Release }},appVersion={{ .Release }}`)
	addStringFlag(cmd, "helm-values-release-template-list", "", "", `list of key value to be replaced in the values.yaml during the release. Example: image.tag={{ .Release }}
	Values: `+templateValues)

	return cmd
}

func helmRelease(pro *Project, flags helmReleaseFlags) {

	if pro.Count() != "0" || len(pro.Tag()) == 0 {
		log.Fatal("Can not created healm release. Missing tag on current commit",
			log.Fields{"version": pro.Version(), "hash": pro.Hash(), "count": pro.Count(), "tag": pro.Tag()})
	}

	// switch back to rc version
	pro.version = pro.rcVersion
	pro.release = pro.rcRelease
	log.Info("Create helm release", log.Fields{"version": pro.Version(), "release": pro.Release()})

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
	updateHelmValues(pro, flags.Helm, flags.ValuesReleaseTemplate)

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
