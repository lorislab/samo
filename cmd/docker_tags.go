package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

func createDockerTagsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tags",
		Short: "Create list of docker image tags",
		Long:  `Create list of docker image tags`,
		Run: func(cmd *cobra.Command, args []string) {

			flags := dockerFlags{}
			readOptions(&flags)
			project := loadProject(flags.Project)
			dockerTagsCmd(project, flags)
		},
		TraverseChildren: true,
	}

	return cmd
}

func dockerTagsCmd(project *Project, flags dockerFlags) {
	var dockerImage string
	tags := dockerTags(dockerImage, project, flags.TagListTemplate)
	fmt.Printf("%s\n", strings.Join(tags, ","))
}
