package cmd

import (
	"github.com/lorislab/samo/project"
	"github.com/spf13/cobra"
)

func createDockerPushCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "push",
		Short: "Push the docker image of the project",
		Long:  `Push the docker image of the project`,
		Run: func(cmd *cobra.Command, args []string) {

			flags := dockerFlags{}
			readOptions(&flags)
			project := loadProject(flags.Project)
			dockerPush(project, flags)
		},
		TraverseChildren: true,
	}

	return cmd
}

func dockerPush(project *project.Project, flags dockerFlags) {
	dockerImage := dockerImage(project, flags.Registry, flags.Group, flags.Repo)
	tags := dockerTags(dockerImage, project, flags)
	dockerImagePush(dockerImage, tags, flags.Project.SkipPush)
}
