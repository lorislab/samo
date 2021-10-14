package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type projectVersionFlags struct {
	Project projectFlags `mapstructure:",squash"`
	Version string       `mapstructure:"version"`
}

func createProjectVersionCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show the project version",
		Long: `Show the project version.

Version types:
  version  current version base on the template 'version-template'. 
           Default template: {{ .Version }}-rc.{{ .Count }}
  release  release/final version of the project`,
		Run: func(cmd *cobra.Command, args []string) {

			flags := projectVersionFlags{}
			readOptions(&flags)
			project := loadProject(flags.Project)
			version := "?"
			switch flags.Version {
			case "version":
				version = project.Version()
			case "release":
				version = project.ReleaseVersion()
			}
			fmt.Printf("%s\n", version)
		},
		TraverseChildren: true,
	}

	addStringFlag(cmd, "version", "", "version", "project version type, one of version | release ")
	return cmd
}
