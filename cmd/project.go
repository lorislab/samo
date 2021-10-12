package cmd

import (
	"github.com/spf13/cobra"
)

type projectFlags struct {
	FirstVersion    string `mapstructure:"first-version"`
	ReleaseMajor    bool   `mapstructure:"release-major"`
	ReleasePatch    bool   `mapstructure:"release-patch"`
	VersionTemplate string `mapstructure:"version-template"`
	SkipPush        bool   `mapstructure:"skip-push"`
}

var versionTemplateInfo = `the version go temmplate string.
values: Tag,Hash,Count,Branch
functions:  trunc <lenght>
For example: {{ .Tag }}-{{ trunc 10 .Hash }}
`

func createProjectCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:              "project",
		Short:            "Project operation",
		Long:             `Tasks for the project. To build, push or release docker and helm artifacts of the project.`,
		TraverseChildren: true,
	}

	addStringFlag(cmd, "first-version", "", "0.0.0", "the first version of the project")
	addBoolFlag(cmd, "release-major", "", false, "create a major release")
	addBoolFlag(cmd, "release-patch", "", false, "create a patch release")
	addStringFlag(cmd, "version-template", "t", "{{ .Tag }}-rc.{{ .Count }}", versionTemplateInfo)
	addBoolFlag(cmd, "skip-push", "", false, "skip git push changes")

	addChildCmd(cmd, createProjectVersionCmd())
	addChildCmd(cmd, createProjectNameCmd())
	addChildCmd(cmd, createProjectReleaseCmd())
	addChildCmd(cmd, createProjectPatchCmd())

	return cmd
}
