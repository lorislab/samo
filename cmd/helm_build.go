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

	addStringFlag(cmd, "source", "", "src/main/helm", "filter project helm chart source directory")
	addBoolFlag(cmd, "filter", "", false, "filter helm resources from soruce to output directory")
	addStringFlag(cmd, "template-func", "", "maven", "template function to replace variables.")
	return cmd
}

func helmBuild(project *Project, flags helmBuildFlags) {

	if _, err := os.Stat(flags.Source); os.IsNotExist(err) {
		log.WithField("input", flags.Source).Fatal("Input helm directory does not exists!")
	}

	// clean helm dir
	healmClean(flags.Helm)
	// add custom helm repo
	healmAddRepo(flags.Helm)
	// update helm repo
	helmRepoUpdate()

	// filter resources to output dir
	buildHelmChart(flags, project)

	// update helm dependencies
	tools.ExecCmd("helm", "dependency", "update", flags.Helm.Dir)
	// package helm chart
	helmPackage(flags.Helm)
}

// Filter filter helm resources
func buildHelmChart(flags helmBuildFlags, pro *Project) {

	templateF := emptyFilter
	if flags.Filter {
		if len(flags.TemplateFuc) == 0 {
			log.WithField("template", flags.TemplateFuc).Warn("Template is empty! Switch back to maven template")
			flags.TemplateFuc = "maven"
		}
		templateF = templateFunc(flags.TemplateFuc)
	}

	// get all files from the input directory
	paths, err := tools.GetAllFilePathsInDirectory(flags.Source)
	if err != nil {
		log.WithField("input", flags.Source).Panic(err)
	}

	// filter helm resources
	for _, path := range paths {
		// load file
		log.WithFields(log.Fields{"file": path}).Debug("Filter file")
		result, err := ioutil.ReadFile(path)
		if err != nil {
			log.Panic(err)
		}
		// replace values in the file
		result = templateF(pro, result)
		// write result to output directory
		tools.WriteBytesToFile(strings.Replace(path, flags.Source, flags.Helm.Dir, -1), result)
	}

}

type tf func(pro *Project, data []byte) []byte

func templateFunc(template string) tf {
	switch template {
	case "maven":
		return func(pro *Project, data []byte) []byte {
			return templateMavenFilter(pro, data)
		}
	default:
		log.WithField("template", template).Fatal("Not supported template!")
	}
	return nil
}

func emptyFilter(pro *Project, data []byte) []byte {
	return data
}

func templateMavenFilter(pro *Project, data []byte) []byte {
	result := regexProjectName.ReplaceAll(data, []byte(pro.Name()))
	result = regexProjectName2.ReplaceAll(result, []byte(pro.Name()))
	result = regexProjectVersion.ReplaceAll(result, []byte(pro.Version()))
	return result
}
