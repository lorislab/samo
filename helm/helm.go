package helm

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/lorislab/samo/project"
	"github.com/lorislab/samo/tools"
	"github.com/lorislab/samo/yaml"
	log "github.com/sirupsen/logrus"
)

type HelmRequest struct {
	Project           project.Project
	Versions          project.Versions
	PushVersions      project.Versions
	Input             string
	Output            string
	Clean             bool
	Template          string
	ChartUpdate       []string
	ValuesUpdate      []string
	Repository        string
	RepositoryURL     string
	Username          string
	Password          string
	SkipPush          bool
	BuildNumberLength int
	BuildNumberPrefix string
	HashLength        int
	BuildFilter       bool
	AddRepo           bool
}

var (
	regexProjectName    = regexp.MustCompile(`\$\{project\.name\}`)
	regexProjectVersion = regexp.MustCompile(`\$\{project\.version\}`)
)

// Build build helm release
func (request HelmRequest) Build() {

	buildVersion := project.BuildVersion(request.Project, request.HashLength, request.BuildNumberLength, request.BuildNumberPrefix).String()

	// clean output directory
	request.clean()

	// add repository
	request.addRepo()

	// update helm dependencies
	tools.ExecCmd("helm", "dependency", "update", request.helmDir())

	// package helm chart
	request.helmPackage()

	// upload helm chart
	request.push(buildVersion)
}

// Release create helm release
func (request HelmRequest) Release() {

	// clean output directory
	request.clean()

	buildVersion := project.BuildVersion(request.Project, request.HashLength, request.BuildNumberLength, request.BuildNumberPrefix).String()
	releaseVersion := project.ReleaseVersion(request.Project).String()

	// download build version
	request.download(buildVersion)

	data := templateData{
		Version:      releaseVersion,
		BuildVersion: buildVersion,
		Project:      request.Project,
	}
	template := template.New("input-helm-chart")
	request.updateFile("Chart.yaml", request.ChartUpdate, template, data)
	request.updateFile("values.yaml", request.ValuesUpdate, template, data)

	// package helm chart
	request.helmPackage()

	// upload helm chart
	request.push(releaseVersion)
}

func (request HelmRequest) updateFile(file string, input []string, template *template.Template, data templateData) {
	tmp := filepath.FromSlash(request.helmDir() + "/" + file)
	log.WithFields(log.Fields{
		"file":   tmp,
		"update": input,
	}).Info("Update helm chart file")
	yaml.ReplaceValueInYaml(tmp, data.template(input, template))
}

type templateData struct {
	Version      string
	BuildVersion string
	Project      project.Project
}

func (r templateData) template(input []string, template *template.Template) map[string]string {

	result := make(map[string]string)
	for _, item := range input {
		items := strings.Split(item, "=")

		t, err := template.Parse(items[1])
		if err != nil {
			log.Panic(err)
		}

		var tpl bytes.Buffer
		err = t.Execute(&tpl, r)
		if err != nil {
			log.Panic(err)
		}

		result[items[0]] = tpl.String()
	}
	return result
}

func (request HelmRequest) helmDir() string {
	return filepath.FromSlash(request.Output + "/" + request.Project.Name())
}

func (request HelmRequest) helmPackage() {
	tools.ExecCmd("helm", "package", request.helmDir())
}

func (request HelmRequest) push(releaseVersion string) {

	// upload helm chart
	if request.SkipPush {
		log.WithFields(log.Fields{
			"repo-url": request.RepositoryURL,
			"version":  releaseVersion,
		}).Info("Skip push release version of the helm chart")
		return
	}

	var command []string
	command = append(command, "-is")
	if len(request.Password) > 0 {
		command = append(command, "-u", `"`+request.Username+`:`+request.Password+`"`)
	}
	command = append(command, request.RepositoryURL, "--upload-file", request.Project.Name()+`-`+releaseVersion+`.tgz`)
	tools.ExecCmd("curl", command...)
}

func (request HelmRequest) addRepo() {
	if !request.AddRepo {
		log.WithFields(log.Fields{
			"repo-url": request.RepositoryURL,
			"repo":     request.Repository,
		}).Info("Skip push release version of the helm chart")
		return
	}

	// add repository
	var command []string
	command = append(command, "repo", "add")
	if len(request.Password) > 0 {
		command = append(command, "--password", request.Password)
	}
	if len(request.Username) > 0 {
		command = append(command, "--username", request.Username)
	}
	command = append(command, request.Repository, request.RepositoryURL)
	tools.ExecCmd("helm", command...)
}

func (request HelmRequest) download(version string) {

	// add repository
	var command []string
	command = append(command, "pull", "--untar", "--untadir", request.Output)
	if len(request.Password) > 0 {
		command = append(command, "--password", request.Password)
	}
	if len(request.Username) > 0 {
		command = append(command, "--username", request.Username)
	}

	url := request.RepositoryURL + "/" + request.Project.Name() + "-" + version + ".tgz"
	command = append(command, url)
	tools.ExecCmd("helm", command...)
}

func (request HelmRequest) clean() {
	// clean output directory
	if request.Clean {
		if _, err := os.Stat(request.Output); !os.IsNotExist(err) {
			err := os.RemoveAll(request.Output)
			if err != nil {
				log.WithField("output", request.Output).Panic(err)
			}
		}
	}
}

type filterData struct {
	Name    string
	Version string
}

// Filter filter helm resources
func (request HelmRequest) Filter() {

	template := request.Template
	if len(template) == 0 {
		log.WithField("template", request.Template).Warn("Template is empty! Switch back to maven template")
		template = "maven"
	}
	templateF := templateFunc(template)

	// output directory output + project.name
	outputDir := request.helmDir()

	if _, err := os.Stat(request.Input); os.IsNotExist(err) {
		log.WithField("input", request.Input).Fatal("Input helm directory does not exists!")
	}

	// clean output directory
	request.clean()

	// get all files from the input directory
	paths, err := tools.GetAllFilePathsInDirectory(request.Input)
	if err != nil {
		log.WithField("input", request.Input).Panic(err)
	}

	version := request.Project.Version()
	filterData := filterData{Name: request.Project.Name(), Version: version}
	for _, path := range paths {
		// load file
		result, err := ioutil.ReadFile(path)
		if err != nil {
			log.Panic(err)
		}
		result = templateF(filterData, result)
		// write result to output directory
		tools.WriteBytesToFile(strings.Replace(path, request.Input, outputDir, -1), result)
	}

}

func templateFunc(template string) func(p filterData, data []byte) []byte {
	switch template {
	case "maven":
		return func(p filterData, data []byte) []byte {
			return templateMavenFilter(p, data)
		}
	default:
		log.WithField("template", template).Fatal("Not supported template!")
	}
	return nil
}

func templateMavenFilter(p filterData, data []byte) []byte {
	result := regexProjectName.ReplaceAll(data, []byte(p.Name))
	result = regexProjectVersion.ReplaceAll(result, []byte(p.Version))
	return result
}
