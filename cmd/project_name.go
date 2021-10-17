package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func createProjectNameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "name",
		Short: "Show the project name",
		Long:  `Tasks to show the maven project name`,
		Run: func(cmd *cobra.Command, args []string) {

			flags := projectFlags{}
			readOptions(&flags)
			project := loadProject(flags)

			fmt.Printf("%s\n", project.Name())
		},
		TraverseChildren: true,
	}
	return cmd
}
