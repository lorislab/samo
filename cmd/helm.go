package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lorislab/samo/tools"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var yamlKeyRegex = regexp.MustCompile(`^"|['"](\w+(?:\.\w+)*)['"]|(\w+)`)

type helmFlags struct {
	Project              projectFlags `mapstructure:",squash"`
	Repo                 string       `mapstructure:"repo"`
	RepoUsername         string       `mapstructure:"repo-username"`
	RepoPassword         string       `mapstructure:"repo-password"`
	RepositoryURL        string       `mapstructure:"repo-url"`
	Clean                bool         `mapstructure:"clean"`
	PushURL              string       `mapstructure:"push-url"`
	PushType             string       `mapstructure:"push-type"`
	Dir                  string       `mapstructure:"dir"`
	ChartFilterTemplate  string       `mapstructure:"chart-template-list"`
	ValuesFilterTemplate string       `mapstructure:"values-template-list"`
}

func createHelmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "helm",
		Short:            "Project helm operation",
		Long:             `Tasks for the helm of the project`,
		TraverseChildren: true,
	}

	addBoolFlag(cmd, "clean", "", false, "clean output directory before filter")
	addStringFlag(cmd, "dir", "", "target/helm", "filter project helm chart output directory")
	addStringFlag(cmd, "repo", "", "", "helm repository name")
	addStringFlag(cmd, "repo-url", "", "", "helm repository URL")
	addStringFlag(cmd, "repo-username", "u", "", "helm repository username")
	addStringFlag(cmd, "repo-password", "p", "", "helm repository password")
	addStringFlag(cmd, "push-url", "", "", "helm repository push URL")
	addStringFlag(cmd, "push-type", "", "harbor", "helm repository push type. Values: upload,harbor")
	addStringFlag(cmd, "chart-template-list", "", "version={{ .Version }},appVersion={{ .Version }},name={{ .Name }}", `list of key value to be replaced in the Chart.yaml
	Values: Name,Hash,Branch,Tag,Count,Version,Release. 
	Example: version={{ .Release }},appVersion={{ .Release }}`)
	addStringFlag(cmd, "values-template-list", "", "", `list of key value to be replaced in the values.yaml Example: image.tag={{ .Version }}
	Values: Hash,Branch,Tag,Count,Version,Release.`)

	addChildCmd(cmd, createHealmBuildCmd())
	addChildCmd(cmd, createHealmPushCmd())
	addChildCmd(cmd, createHealmReleaseCmd())
	return cmd
}

func helmPackage(project *Project, flags helmFlags) {
	tools.ExecCmd("helm", "package", helmDir(project, flags))
}

func healmClean(flags helmFlags) {
	// clean output directory
	if !flags.Clean {
		log.Debug("Helm clean disabled.")
		return
	}
	if _, err := os.Stat(flags.Dir); !os.IsNotExist(err) {
		log.WithField("dir", flags.Dir).Debug("Clean directory")
		err := os.RemoveAll(flags.Dir)
		if err != nil {
			log.WithField("output", flags.Dir).Panic(err)
		}
	}
}

func healmAddRepo(flags helmFlags) {
	if len(flags.Repo) == 0 {
		return
	}

	// add repository
	var command []string
	command = append(command, "repo", "add")
	if len(flags.RepoPassword) > 0 {
		command = append(command, "--password", flags.RepoPassword)
	}
	if len(flags.RepoUsername) > 0 {
		command = append(command, "--username", flags.RepoUsername)
	}
	command = append(command, flags.Repo, flags.RepositoryURL)
	tools.ExecCmd("helm", command...)
}

func helmRepoUpdate() {
	tools.ExecCmd("helm", "repo", "update")
}

func helmPush(version string, project *Project, flags helmFlags) {

	if len(flags.PushURL) == 0 {
		log.WithFields(log.Fields{"push-url": flags.PushURL, "version": version}).Fatal("Flag --push-url is mandatory!")
	}

	// upload helm chart
	if flags.Project.SkipPush {
		log.WithFields(log.Fields{"push-url": flags.PushURL, "version": version}).Info("Skip push release version of the helm chart")
		return
	}

	filename := project.Name() + `-` + version + `.tgz`
	if !tools.Exists(filename) {
		log.WithField("helm-file", filename).Fatal("Helm package file does not exists!")
	}

	var command []string
	command = append(command, "-fis", "--show-error")
	if len(flags.RepoPassword) > 0 {
		command = append(command, "-u", flags.RepoUsername+`:`+flags.RepoPassword)
	}

	switch flags.PushType {
	case "upload":
		command = append(command, flags.PushURL, "--upload-file", filename)
	case "harbor":
		command = append(command, "-F", `"chart=@`+filename+`"`, flags.PushURL)
	default:
		log.WithField("push-type", flags.PushType).Fatal("Not supported helm push type")
	}

	tools.ExecCmd("curl", command...)
}

// update helm version, appversion, annotations/labels in Chart.yaml
func updateHelmValues(project *Project, flags helmFlags) {
	if len(flags.ValuesFilterTemplate) < 1 {
		return
	}
	data := map[string]string{}
	t := templateToMap(flags.ValuesFilterTemplate, project)
	for k, v := range t {
		data[k] = v
	}
	if len(data) > 0 {
		file := filepath.FromSlash(helmDir(project, flags) + "/values.yaml")
		replaceValueInYaml(file, data)
	}
}

// update helm version, appversion, annotations/labels in Chart.yaml
func updateHelmChart(project *Project, flags helmFlags, chartTemplate string) {
	data := map[string]string{}

	if !flags.Project.SkipLabels {
		data[`annotations."samo.project.hash"`] = project.Hash()
		data[`annotations."samo.project.version"`] = project.Version()
	}
	if len(flags.Project.LabelTemplate) > 0 {
		t := templateToMap(flags.Project.LabelTemplate, project)
		for k, v := range t {
			data[`annotations."`+k+`"`] = v
		}
	}
	if len(chartTemplate) > 0 {
		t := templateToMap(chartTemplate, project)
		for k, v := range t {
			data[k] = v
		}
	}
	if len(data) < 1 {
		return
	}
	file := filepath.FromSlash(helmDir(project, flags) + "/Chart.yaml")
	replaceValueInYaml(file, data)
}

func helmDir(project *Project, flags helmFlags) string {
	return flags.Dir + "/" + project.name
}

func templateToMap(template string, data interface{}) map[string]string {
	r := map[string]string{}
	labelTemplate := tools.Template(data, template)
	labels := strings.Split(labelTemplate, ",")
	for _, label := range labels {
		v := strings.SplitN(label, "=", 2)
		r[v[0]] = v[1]
	}
	return r
}

func replaceValueInYaml(filename string, data map[string]string) {

	obj := make(map[interface{}]interface{})

	fileBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Panic(err)
	}
	err = yaml.Unmarshal(fileBytes, &obj)
	if err != nil {
		log.Panic(err)
	}
	for k, v := range data {
		replace(obj, k, v)
	}

	fileBytes, err = yaml.Marshal(&obj)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	err = ioutil.WriteFile(filename, fileBytes, 0666)
	if err != nil {
		log.Panic(err)
	}
	log.WithFields(log.Fields{"file": filename}).Info("Update file")
}

func replace(obj map[interface{}]interface{}, k string, v string) {
	// keys := strings.Split(k, ".")

	keys := yamlKeyRegex.FindAllString(k, -1)
	// keys := yamlKeyRegex.FindAllStringSubmatch(k, -1)

	var tmp interface{}
	size := len(keys)

	tmp = obj
	for i := 0; i < size-1; i++ {
		key := keys[i]
		key = strings.TrimSuffix(strings.TrimPrefix(key, `"`), `"`)
		a := tmp.(map[interface{}]interface{})[key]
		if a == nil {
			a = map[interface{}]interface{}{}
			tmp.(map[interface{}]interface{})[key] = a
		}
		tmp = a
	}
	key := keys[size-1]
	key = strings.TrimSuffix(strings.TrimPrefix(key, `"`), `"`)
	tmp.(map[interface{}]interface{})[key] = v
}
