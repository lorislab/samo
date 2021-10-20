package cmd

import (
	"github.com/lorislab/samo/log"
	"github.com/lorislab/samo/tools"
	"github.com/spf13/cobra"
)

type projectReleaseFlags struct {
	Project         projectFlags `mapstructure:",squash"`
	MessageTemplate string       `mapstructure:"message-template"`
	TagTemplate     string       `mapstructure:"tag-template"`
}

func createProjectReleaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "release",
		Short: "Create release of the current project and state",
		Long:  `Create release of the current project and state`,
		Run: func(cmd *cobra.Command, args []string) {
			flags := projectReleaseFlags{}
			readOptions(&flags)
			project := loadProject(flags.Project)
			release(project, flags)
		},
		TraverseChildren: true,
	}
	addStringFlag(cmd, "message-template", "", "{{ .Release }}", `the annotated tag message template.
	Values: `+templateValues)
	addStringFlag(cmd, "tag-template", "", "{{ .Release }}", `the release tag template. 
	Values: `+templateValues)

	return cmd
}

// CreateRelease create project release
func release(pro *Project, flags projectReleaseFlags) {
	tag := tools.Template(pro, flags.TagTemplate)
	msg := tools.Template(pro, flags.MessageTemplate)
	tools.Git("tag", "-a", tag, "-m", msg)

	// push project to remote repository
	if flags.Project.SkipPush {
		log.Info("Skip git push for project release", log.F("version", tag))
	} else {
		tools.Git("push", "--tags")
	}
	log.Info("New release created.", log.F("version", tag))
}
