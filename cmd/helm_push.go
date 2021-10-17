package cmd

import (
	"github.com/spf13/cobra"
)

func createHealmPushCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "push",
		Short: "Push helm chart",
		Long:  `Push helm chart to the helm repository`,
		Run: func(cmd *cobra.Command, args []string) {

			flags := helmFlags{}
			readOptions(&flags)
			project := loadProject(flags.Project)
			helmPush(project.Version(), project, flags)
		},
		TraverseChildren: true,
	}

	return cmd
}
