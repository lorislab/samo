package cmd

import (
	"fmt"
	"github.com/lorislab/samo/tools"
	"github.com/spf13/cobra"
)

type dockerAnnotationFlags struct {
	Docker             dockerFlags `mapstructure:",squash"`
	IterableTemplate   string      `mapstructure:"docker-annotation-template-type"`
	AnnotationTemplate string      `mapstructure:"docker-annotation-template"`
}

var defaultTemplateOnLine = "{{ range $index, $item := .Annotations }}{{if $index}},{{end}}annotation-index.{{ $item.Key }}={{ $item.Value}}{{ end }}"
var defaultTemplateMultiLines = "{{ .Key }}={{ .Value }}"

func createDockerAnnotationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "annotations",
		Short: "Create list of docker image annotations",
		Long:  `Create list of docker image annotations`,
		Run: func(cmd *cobra.Command, args []string) {

			flags := dockerAnnotationFlags{}
			readOptions(&flags)
			project := loadProject(flags.Docker.Project)
			dockerAnnotationsCmd(project, flags)
		},
		TraverseChildren: true,
	}

	addStringFlag(cmd, "docker-annotation-template-type", "", "one-line", "use the template [ one-line | multi-lines ]")
	addStringFlag(cmd, "docker-annotation-template", "", "",
		`Default one-line annotation template:
  `+defaultTemplateOnLine+`
Default multi-lines annotation template:
  `+defaultTemplateMultiLines+`
`)

	return cmd
}

func dockerAnnotationsCmd(project *Project, flags dockerAnnotationFlags) {

	annotations := dockerLabels(project, flags.Docker.Project.SkipLabels, flags.Docker.SkipOpenContainersLabels, flags.Docker.Project.LabelTemplate)

	var template = flags.AnnotationTemplate
	if flags.IterableTemplate == "one-line" {
		if len(template) == 0 {
			template = defaultTemplateOnLine
		}
		templateOneLine(annotations, template)
	} else {
		if len(template) == 0 {
			template = defaultTemplateMultiLines
		}
		templateMultiLines(annotations, template)
	}
}

type A struct {
	Annotations []I
}

func templateOneLine(annotations map[string]string, template string) {

	var tmp []I
	for k, v := range annotations {
		tmp = append(tmp, I{k, v})
	}
	output := tools.Template(A{tmp}, template)
	fmt.Printf("%s\n", output)
}

type I struct {
	Key   string
	Value string
}

func templateMultiLines(annotations map[string]string, template string) {
	var output []string
	for k, v := range annotations {
		label := tools.Template(I{k, v}, template)
		output = append(output, label)
	}

	// print labels
	for _, label := range output {
		fmt.Printf("%s\n", label)
	}
}
