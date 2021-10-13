package cmd

import (
	"github.com/lorislab/samo/project"
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
	addStringFlag(cmd, "message-template", "", "{{ .Project.Version }}", `the annotated tag message template.
	Values: Project.Hash,Project.Branch,Project.Tag,Project.Count,Project.Version,Project.Release.`)
	addStringFlag(cmd, "tag-template", "", "{{ .Project.Version }}", `the release tag template. 
	Values: Project.Hash,Project.Branch,Project.Tag,Project.Count,Project.Version,Project.Release.`)

	return cmd
}

// CreateRelease create project release
func release(pro *project.Project, flags projectReleaseFlags) {

	data := struct {
		Project *project.Project
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
