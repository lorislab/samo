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
	Project       project.Project
	Versions      project.Versions
	Input         string
	Output        string
	Clean         bool
	Template      string
	FilterChart   []string
	FilterValues  []string
	Repository    string
	RepositoryURL string
	Username      string
	Password      string
	SkipPush      bool
	Filter        bool
	AddRepo       bool
	UpdateVersion bool
}

var (
	regexProjectName    = regexp.MustCompile(`\$\{project\.name\}`)
	regexProjectName2   = regexp.MustCompile(`\$\{project\.artifactId\}`)
	regexProjectVersion = regexp.MustCompile(`\$\{project\.version\}`)
)

// Build build helm release
func (request HelmRequest) Build() {

	// clean output directory
	request.clean()

	version := request.Versions.Unique()

	// filter resources to output dir
	request.filter(version)

	// add repository
	request.addRepo()

	// update helm dependencies
	tools.ExecCmd("helm", "dependency", "update", request.helmDir())

	// update version
	if request.UpdateVersion {
		data := templateData{
			Version:      version,
			BuildVersion: request.Project.Version(),
			Project:      request.Project,
		}
		request.updateVersion(data)
	} else {
		log.Debug("Skip update version.")
	}

	// package helm chart
	request.helmPackage()
}

// Build build helm release
func (request HelmRequest) Push() {

	// version
	version := request.Versions.Unique()

	// upload helm chart
	request.push(version)
}

// Release create helm release
func (request HelmRequest) Release() {

	// clean output directory
	request.clean()

	// update repositories
	request.repoUpdate()

	// download build version
	buildVersion := request.Versions.BuildVersion()
	request.download(buildVersion)

	releaseVersion := request.Versions.ReleaseVersion()
	data := templateData{
		Version:      releaseVersion,
		BuildVersion: buildVersion,
		Project:      request.Project,
	}
	request.updateVersion(data)

	// package helm chart
	request.helmPackage()

	// upload helm chart
	request.push(releaseVersion)
}

// update helm chart with a release version
func (request HelmRequest) updateVersion(data templateData) {
	template := template.New("input-helm-chart")
	if len(request.FilterChart) > 0 {
		request.updateFile("Chart.yaml", request.FilterChart, template, data)
	}
	if len(request.FilterValues) > 0 {
		request.updateFile("values.yaml", request.FilterValues, template, data)
	}
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

	filename := request.Project.Name() + `-` + releaseVersion + `.tgz`
	if !tools.Exists(filename) {
		log.WithField("helm-file", filename).Fatal("Helm package file does not exists!")
	}

	var command []string
	command = append(command, "-fis", "--show-error")
	if len(request.Password) > 0 {
		command = append(command, "-u", request.Username+`:`+request.Password)
	}
	command = append(command, request.RepositoryURL, "--upload-file", filename)
	tools.ExecCmd("curl", command...)
}

func (request HelmRequest) addRepo() {
	if !request.AddRepo {
		log.WithFields(log.Fields{
			"repo-url": request.RepositoryURL,
			"repo":     request.Repository,
		}).Debug("Skip add helm repository")
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

	request.repoUpdate()
}

func (request HelmRequest) repoUpdate() {
	tools.ExecCmd("helm", "repo", "update")
}

func (request HelmRequest) download(version string) {

	// add repository
	var command []string
	command = append(command, "pull", "--untar", "--untardir", request.Output)
	if len(request.Password) > 0 {
		command = append(command, "--password", request.Password)
	}
	if len(request.Username) > 0 {
		command = append(command, "--username", request.Username)
	}

	url := request.RepositoryURL
	if !strings.HasSuffix(url, "/") {
		url = url + "/"
	}
	url = url + request.Project.Name() + "-" + version + ".tgz"
	command = append(command, url)
	tools.ExecCmd("helm", command...)
}

func (request HelmRequest) clean() {
	// clean output directory
	if !request.Clean {
		log.Debug("Helm clean disabled.")
		return
	}

	if _, err := os.Stat(request.Output); !os.IsNotExist(err) {
		log.WithField("dir", request.Output).Debug("Clean directory")
		err := os.RemoveAll(request.Output)
		if err != nil {
			log.WithField("output", request.Output).Panic(err)
		}
	}
}

type filterData struct {
	Name    string
	Version string
}

// Filter filter helm resources
func (request HelmRequest) filter(version string) {

	// clean output directory
	if !request.Filter {
		log.Debug("Helm filter disabled.")
		return
	}
	if len(request.Template) <= 0 {
		log.Fatal("Missing template for the helm chart filtering!")
	}

	template := request.Template
	if len(template) == 0 {
		log.WithField("template", request.Template).Warn("Template is empty! Switch back to maven template")
		template = "maven"
	}
	templateF := templateFunc(template)

	if _, err := os.Stat(request.Input); os.IsNotExist(err) {
		log.WithField("input", request.Input).Fatal("Input helm directory does not exists!")
	}

	// get all files from the input directory
	paths, err := tools.GetAllFilePathsInDirectory(request.Input)
	if err != nil {
		log.WithField("input", request.Input).Panic(err)
	}

	// filter helm resources
	outputDir := request.helmDir()
	filterData := filterData{Name: request.Project.Name(), Version: version}
	for _, path := range paths {
		// load file
		log.WithFields(log.Fields{
			"file": path,
			"data": filterData,
		}).Debug("Filter file")
		result, err := ioutil.ReadFile(path)
		if err != nil {
			log.Panic(err)
		}
		// replace values in the file
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
	result = regexProjectName2.ReplaceAll(result, []byte(p.Name))
	result = regexProjectVersion.ReplaceAll(result, []byte(p.Version))
	return result
}
