package cmd

import (
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/lorislab/samo/tools"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	regexProjectName    = regexp.MustCompile(`\$\{project\.name\}`)
	regexProjectName2   = regexp.MustCompile(`\$\{project\.artifactId\}`)
	regexProjectVersion = regexp.MustCompile(`\$\{project\.version\}`)
)

type helmBuildFlags struct {
	Helm        helmFlags `mapstructure:",squash"`
	Source      string    `mapstructure:"source"`
	Filter      bool      `mapstructure:"filter"`
	TemplateFuc string    `mapstructure:"template-func"`
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

	addStringFlag(cmd, "source", "", "", "filter project helm chart source directory")
	addBoolFlag(cmd, "filter", "", false, "filter helm resources from soruce to output directory")
	addStringFlag(cmd, "template-func", "", "no-filter", "template function to replace variables. Funnctions: no-filter,maven")
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

	templateF := templateFunc(flags.TemplateFuc)

	// get all files from the input directory
	if len(flags.Source) > 0 {

		if _, err := os.Stat(flags.Source); os.IsNotExist(err) {
			log.WithField("source", flags.Source).Fatal("Source helm directory does not exists!")
		}

		paths, err := tools.GetAllFilePathsInDirectory(flags.Source)
		if err != nil {
			log.WithField("source", flags.Source).Panic(err)
		}

		// filter helm resources
		for _, path := range paths {
			// load file
			log.WithFields(log.Fields{"file": path}).Info("Filter file")
			result, err := ioutil.ReadFile(path)
			if err != nil {
				log.Panic(err)
			}
			// replace values in the file
			result = templateF(pro, result)
			// write result to output directory
			out := strings.Replace(path, flags.Source, flags.Helm.Dir, -1)
			tools.WriteBytesToFile(out, result)
			log.WithFields(log.Fields{"file": out}).Info("Create file")
		}
	}
}

type tf func(pro *Project, data []byte) []byte

func templateFunc(template string) tf {
	switch template {
	case "", "no-filter":
		return noFilter
	case "maven":
		return templateMavenFilter
	default:
		log.WithField("template", template).Warn("Not supported template! Switch back to no filter.")
	}
	return noFilter
}

func noFilter(pro *Project, data []byte) []byte {
	return data
}

func templateMavenFilter(pro *Project, data []byte) []byte {
	result := regexProjectName.ReplaceAll(data, []byte(pro.Name()))
	result = regexProjectName2.ReplaceAll(result, []byte(pro.Name()))
	result = regexProjectVersion.ReplaceAll(result, []byte(pro.Version()))
	return result
}
