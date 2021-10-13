package cmd

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/lorislab/samo/tools"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type helmReleaseFlags struct {
	Helm                 helmFlags `mapstructure:",squash"`
	ChartFilterTemplate  string    `mapstructure:"chart-filter-template"`
	ValuesFilterTemplate string    `mapstructure:"values-filter-template"`
}

func createHealmReleaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "release",
		Short: "Release helm chart",
		Long:  `Download version of the helm chart and create final version`,
		Run: func(cmd *cobra.Command, args []string) {
			flags := helmReleaseFlags{}

			// remove all tags from current commit
			tools.GitRemoveAllTagsForCurrentCommit()

			readOptions(&flags)
			project := loadProject(flags.Helm.Project)
			helmRelease(project, flags)
		},
		TraverseChildren: true,
	}

	addStringFlag(cmd, "chart-filter-template", "", "version={{ .Release }},appVersion={{ .Release }}", `list of key value to be replaced in the Chart.yaml
	Values: Hash,Branch,Tag,Count,Version,Release. 
	Example: version={{ .Version }},appVersion={{ .Hash }}
	`)
	addStringFlag(cmd, "values-filter-template", "", "", `list of key value to be replaced in the values.yaml Example: image.tag={{ .Release }}
	Values: Hash,Branch,Tag,Count,Version,Release. 
	`)
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
	helmDownload(pro, flags)
	// update version to release version
	updateVersion(pro, flags)
	// package helm chart
	helmPackage(flags.Helm)
	// upload helm chart with release version
	helmPush(pro.ReleaseVersion(), pro, flags.Helm)
}

func updateVersion(pro *Project, flags helmReleaseFlags) {
	if len(flags.ChartFilterTemplate) > 0 {
		updateFile("Chart.yaml", flags.ChartFilterTemplate, flags.Helm, pro)
	}
	if len(flags.ValuesFilterTemplate) > 0 {
		updateFile("values.yaml", flags.ValuesFilterTemplate, flags.Helm, pro)
	}
}

func updateFile(filename, template string, flags helmFlags, pro *Project) {
	tmp := tools.Template(pro, template)
	items := strings.Split(tmp, ",")

	data := map[string]string{}
	for _, item := range items {
		kv := strings.Split(item, "=")
		data[kv[0]] = kv[1]
	}

	file := filepath.FromSlash(flags.Dir + "/" + filename)
	log.WithFields(log.Fields{"file": file}).Info("Update file")
	replaceValueInYaml(file, data)
}

func helmDownload(project *Project, flags helmReleaseFlags) {
	var command []string
	command = append(command, "pull")
	command = append(command, flags.Helm.Repo+"/"+project.Name())
	command = append(command, "--version", project.Version())
	command = append(command, "--untar", "--untardir", flags.Helm.Dir)
	tools.ExecCmd("helm", command...)
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
}

func replace(obj map[interface{}]interface{}, k string, v string) {
	keys := strings.Split(k, ".")
	var tmp interface{}
	size := len(keys)

	tmp = obj
	for i := 0; i < size-1; i++ {
		key := keys[i]
		a := tmp.(map[interface{}]interface{})[key]
		if a == nil {
			a = map[interface{}]interface{}{}
			tmp.(map[interface{}]interface{})[key] = a
		}
		tmp = a
	}
	tmp.(map[interface{}]interface{})[keys[size-1]] = v
}
