package cmd

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/lorislab/samo/log"
	"github.com/lorislab/samo/tools"
	"github.com/spf13/cobra"
)

type helmBuildFlags struct {
	Helm   helmFlags `mapstructure:",squash"`
	Source string    `mapstructure:"helm-source"`
}

func createHealmBuildCmd() *cobra.Command {
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

	addStringFlag(cmd, "helm-source", "", "", "filter project helm chart source directory")
	return cmd
}

func helmBuild(project *Project, flags helmBuildFlags) {

	// clean helm dir
	healmClean(flags.Helm)
	// add custom helm repo
	healmAddRepo(flags.Helm)
	// update helm repo
	helmRepoUpdate()

	// filter resources to output dir
	buildHelmChart(flags, project)

	updateHelmChart(project, flags.Helm, flags.Helm.ChartFilterTemplate)
	updateHelmValues(project, flags.Helm)

	// update helm dependencies
	tools.ExecCmd("helm", "dependency", "update", helmDir(project, flags.Helm))

	// package helm chart
	helmPackage(project, flags.Helm)
}

// Filter filter helm resources
func buildHelmChart(flags helmBuildFlags, pro *Project) {

	// get all files from the input directory
	if len(flags.Source) > 0 {

		if _, err := os.Stat(flags.Source); os.IsNotExist(err) {
			log.Fatal("Source helm directory does not exists!", log.F("source", flags.Source))
		}

		paths, err := tools.GetAllFilePathsInDirectory(flags.Source)
		if err != nil {
			log.Fatal("error read helm source directory", log.F("source", flags.Source).E(err))
		}

		for _, path := range paths {
			// load file
			result, err := ioutil.ReadFile(path)
			if err != nil {
				log.Fatal("error read file", log.F("file", path).E(err))
			}
			// write result to output directory
			out := strings.Replace(path, flags.Source, flags.Helm.Dir, -1)
			tools.WriteBytesToFile(out, result)
			log.Debug("Copy file", log.F("file", out))
		}
	}
}
