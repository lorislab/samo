package cmd

import (
	"os"
	"strings"

	"github.com/lorislab/samo/log"
	"github.com/lorislab/samo/tools"
	"github.com/spf13/cobra"
)

type helmBuildFlags struct {
	Helm                 helmFlags `mapstructure:",squash"`
	Source               string    `mapstructure:"helm-source-dir"`
	Copy                 bool      `mapstructure:"helm-source-copy"`
	ChartFilterTemplate  string    `mapstructure:"helm-chart-template-list"`
	ValuesFilterTemplate string    `mapstructure:"helm-values-template-list"`
}

func createHelmBuildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build helm chart",
		Long:  `Helm build helm chart`,
		Run: func(cmd *cobra.Command, args []string) {
			flags := helmBuildFlags{}
			readOptions(&flags)
			project := loadProject(flags.Helm.Project)
			helmBuild(project, flags)
		},
		TraverseChildren: true,
	}

	addStringFlag(cmd, "helm-source-dir", "", "", "project helm chart source directory")
	addBoolFlag(cmd, "helm-source-copy", "", false, "copy helm source to helm directory")
	addStringFlag(cmd, "helm-chart-template-list", "", "version={{ .Version }},appVersion={{ .Version }},name={{ .Name }}", `list of key value to be replaced in the Chart.yaml
	Values: `+templateValues+`
	Example: version={{ .Release }},appVersion={{ .Release }}`)
	addStringFlag(cmd, "helm-values-template-list", "", "", `list of key value to be replaced in the values.yaml Example: image.tag={{ .Version }}
	Values: `+templateValues)

	return cmd
}

func helmBuild(project *Project, flags helmBuildFlags) {

	// clean helm dir
	helmClean(flags.Helm)
	// add and update custom helm repo
	helmAddRepo(flags.Helm)

	// filter resources to output dir
	buildHelmChart(flags, project)

	updateHelmChart(project, flags.Helm, flags.ChartFilterTemplate)
	updateHelmValues(project, flags.Helm, flags.ValuesFilterTemplate)

	// update helm dependencies
	tools.ExecCmd("helm", "dependency", "update", helmDir(project, flags.Helm))

	// package helm chart
	helmPackage(project, flags.Helm)
}

// Filter filter helm resources
func buildHelmChart(flags helmBuildFlags, pro *Project) {
	if !flags.Copy {
		log.Debug("helm chart copy is disabled")
		return
	}
	if len(flags.Source) < 1 {
		log.Debug("no helm chart source directory configured")
		return
	}
	// get all files from the input directory
	if _, err := os.Stat(flags.Source); os.IsNotExist(err) {
		log.Fatal("source helm directory does not exists!", log.F("source", flags.Source))
	}

	paths, err := tools.GetAllFilePathsInDirectory(flags.Source)
	if err != nil {
		log.Fatal("error read helm source directory", log.F("source", flags.Source).E(err))
	}

	for _, path := range paths {
		// load file
		result, err := os.ReadFile(path)
		if err != nil {
			log.Fatal("error read file", log.F("file", path).E(err))
		}
		// write result to output directory
		out := strings.Replace(path, flags.Source, flags.Helm.Dir+"/"+pro.name, -1)
		tools.WriteBytesToFile(out, result)
		log.Debug("Copy file", log.F("out", out).F("in", path))
	}
}
