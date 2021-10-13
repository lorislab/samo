package cmd

import (
	"github.com/lorislab/samo/tools"
	log "github.com/sirupsen/logrus"
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
	addStringFlag(cmd, "message-template", "", "{{ .Version }}", `the annotated tag message template.
	Values: Hash,Branch,Tag,Count,Version,Release.`)
	addStringFlag(cmd, "tag-template", "", "{{ .Version }}", `the release tag template. 
	Values: Hash,Project.Branch,Tag,Count,Version,Release.`)

	return cmd
}

// CreateRelease create project release
func release(pro *Project, flags projectReleaseFlags) {

	data := struct {
		Project *Project
	}{
		Project: pro,
	}
	tag := tools.Template(data, flags.TagTemplate)
	msg := tools.Template(data, flags.MessageTemplate)
	tools.Git("tag", "-a", tag, "-m", msg)

	// push project to remote repository
	if flags.Project.SkipPush {
		log.WithField("tag", tag).Info("Skip git push for project release")
	} else {
		tools.Git("push", "--tags")
	}
	log.WithField("version", tag).Info("New release created.")
}
