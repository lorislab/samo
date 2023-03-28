package cmd

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/lorislab/samo/log"
	"github.com/lorislab/samo/tools"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var yamlKeyRegex = regexp.MustCompile(`^"|['"](\w+(?:\.\w+)*)['"]|(\w+)`)

type helmFlags struct {
	Project       projectFlags `mapstructure:",squash"`
	Repo          string       `mapstructure:"helm-repo"`
	RepoUsername  string       `mapstructure:"helm-repo-username" yaml:"-"`
	RepoPassword  string       `mapstructure:"helm-repo-password" yaml:"-"`
	RepositoryURL string       `mapstructure:"helm-repo-url"`
	Clean         bool         `mapstructure:"helm-clean"`
	PushURL       string       `mapstructure:"helm-push-url"`
	PushType      string       `mapstructure:"helm-push-type"`
	Dir           string       `mapstructure:"helm-dir"`
	Registry      string       `mapstructure:"helm-registry"`
	AbsoluteDir   bool         `mapstructure:"helm-absolute-dir"`
}

func createHelmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "helm",
		Short:            "Project helm operation",
		Long:             `Tasks for the helm of the project`,
		TraverseChildren: true,
	}

	addBoolFlag(cmd, "helm-clean", "", false, "clean output directory before filter")
	addStringFlag(cmd, "helm-dir", "", "target/helm", "filter project helm chart output directory")
	addStringFlag(cmd, "helm-repo", "", "", "helm repository name [deprecated]")
	addStringFlag(cmd, "helm-repo-url", "", "", "helm repository URL [deprecated]")
	addStringFlag(cmd, "helm-repo-username", "u", "", "helm repository username [deprecated]")
	addStringFlag(cmd, "helm-repo-password", "p", "", "helm repository password [deprecated]")
	addStringFlag(cmd, "helm-push-url", "", "", "helm repository push URL [deprecated]")
	addStringFlag(cmd, "helm-push-type", "", "harbor", "helm repository push type. Values: upload,harbor [deprecated]")
	addStringFlag(cmd, "helm-registry", "", "", "helm OCI registry")
	addBoolFlag(cmd, "helm-absolute-dir", "", false, "helm chart absolute directory (skip add project name in path)")

	addChildCmd(cmd, createHelmBuildCmd())
	addChildCmd(cmd, createHelmPushCmd())
	addChildCmd(cmd, createHelmReleaseCmd())
	return cmd
}

func helmPackage(project *Project, flags helmFlags) {
	tools.ExecCmd("helm", "package", helmDir(project, flags))
}

func helmClean(flags helmFlags) {
	// clean output directory
	if !flags.Clean {
		log.Debug("Helm clean disabled.")
		return
	}
	if _, err := os.Stat(flags.Dir); !os.IsNotExist(err) {
		log.Debug("Clean directory", log.F("dir", flags.Dir))
		err := os.RemoveAll(flags.Dir)
		if err != nil {
			log.Panic("error delete directory", log.F("output", flags.Dir).E(err))
		}
	}
}

// deprecated
func helmAddRepo(flags helmFlags) {
	if len(flags.Repo) == 0 {
		return
	}

	// add repository
	var command []string
	var exclude []int
	command = append(command, "repo", "add")
	if len(flags.RepoPassword) > 0 {
		command = append(command, "--password", flags.RepoPassword)
		exclude = append(exclude, len(command)-1)
	}
	if len(flags.RepoUsername) > 0 {
		command = append(command, "--username", flags.RepoUsername)
		exclude = append(exclude, len(command)-1)
	}
	command = append(command, flags.Repo, flags.RepositoryURL)
	tools.ExecCmdAdv(exclude, "helm", command...)

	// update index of the added repository
	tools.ExecCmd("helm", "repo", "update")
}

func helmPush(version string, project *Project, flags helmFlags) {

	filename := project.Name() + `-` + version + `.tgz`
	if !tools.Exists(filename) {
		log.Fatal("Helm package file does not exists!", log.F("helm-file", filename))
	}

	// upload helm chart
	if flags.Project.SkipPush {
		log.Info("Skip push release version of the helm chart", log.Fields{"push-registry": flags.Registry, "version": version, "push-url": flags.PushURL})
		return
	}

	// push helm repository
	if len(flags.Registry) == 0 {
		helmPushRepository(filename, version, project, flags)
		return
	}

	var command []string
	var exclude []int

	command = append(command, "push")
	command = append(command, filename)
	command = append(command, flags.Registry)
	tools.ExecCmdAdv(exclude, "helm", command...)
}

// deprecated
func helmPushRepository(filename string, version string, project *Project, flags helmFlags) {

	var command []string
	var exclude []int

	if len(flags.PushURL) == 0 {
		log.Fatal("Flag --helm-push-url is mandatory!", log.Fields{"--helm-push-url": flags.PushURL, "version": version})
	}

	command = append(command, "-fis", "--show-error")
	if len(flags.RepoPassword) > 0 {
		command = append(command, "-u", flags.RepoUsername+`:`+flags.RepoPassword)
		exclude = append(exclude, len(command)-1)
	}

	switch flags.PushType {
	case "upload":
		command = append(command, flags.PushURL, "--upload-file", filename)
	case "harbor":
		command = append(command, "-F", `chart=@`+filename, flags.PushURL)
	default:
		log.Fatal("Not supported helm push type", log.F("push-type", flags.PushType))
	}

	tools.ExecCmdAdv(exclude, "curl", command...)
}

// update helm version, app-version, annotations/labels in Chart.yaml
func updateHelmValues(project *Project, flags helmFlags, valuesTemplate string) {
	if len(valuesTemplate) < 1 {
		return
	}
	data := map[string]string{}
	t := templateToMap(valuesTemplate, project)
	for k, v := range t {
		data[k] = v
	}
	if len(data) > 0 {
		file := filepath.FromSlash(helmDir(project, flags) + "/values.yaml")
		replaceValueInYaml(file, data)
	}
}

// update helm version, app version, annotations/labels in Chart.yaml
func updateHelmChart(project *Project, flags helmFlags, chartTemplate string) {
	data := map[string]string{}

	if !flags.Project.SkipLabels {
		data[`annotations."samo.project.revision"`] = project.Hash()
		data[`annotations."samo.project.version"`] = project.Version()
		data[`annotations."samo.project.created"`] = time.Now().Format(time.RFC3339)
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
	if flags.AbsoluteDir {
		return flags.Dir
	}
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

	if !tools.Exists(filename) {
		log.Fatal("Helm yaml file does not exists!", log.F("file", filename))
	}

	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		log.Panic("error read file", log.E(err).F("file", filename))
	}
	err = yaml.Unmarshal(fileBytes, &obj)
	if err != nil {
		log.Panic("error unmarshal file", log.E(err).F("file", filename))
	}
	for k, v := range data {
		replace(obj, k, v)
	}

	fileBytes, err = yaml.Marshal(&obj)
	if err != nil {
		log.Fatal("error marshal file", log.E(err).F("file", filename))
	}

	err = os.WriteFile(filename, fileBytes, 0666)
	if err != nil {
		log.Panic("error write file", log.E(err).F("file", filename))
	}
	log.Info("Update file", log.F("file", filename))
}

func replace(obj map[interface{}]interface{}, k string, v string) {
	keys := yamlKeyRegex.FindAllString(k, -1)

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
