package cmd

import (
	"fmt"
	"github.com/lorislab/samo/tools"
	"github.com/spf13/cobra"
)

type dockerLabelFlags struct {
	Docker        dockerFlags `mapstructure:",squash"`
	LabelTemplate string      `mapstructure:"docker-label-template"`
}

func createDockerLabelsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "labels",
		Short: "Create list of docker image labels",
		Long:  `Create list of docker image labels`,
		Run: func(cmd *cobra.Command, args []string) {

			flags := dockerLabelFlags{}
			readOptions(&flags)
			project := loadProject(flags.Docker.Project)
			dockerLabelsCmd(project, flags)
		},
		TraverseChildren: true,
	}

	addStringFlag(cmd, "docker-label-template", "", "{{ .Key }}={{ .Value }}", "Label template. Values: Key,Value")

	return cmd
}

type T struct {
	Key   string
	Value string
}

func dockerLabelsCmd(project *Project, flags dockerLabelFlags) {

	labels := dockerLabels(project, flags.Docker.Project.SkipLabels, flags.Docker.SkipOpenContainersLabels, flags.Docker.Project.LabelTemplate)
	var output []string

	for k, v := range labels {
		label := tools.Template(T{k, v}, flags.LabelTemplate)
		output = append(output, label)
	}

	// print labels
	for _, label := range output {
		fmt.Printf("%s\n", label)
	}

}
