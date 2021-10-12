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
			Release(project, flags)
		},
		TraverseChildren: true,
	}
	addStringFlag(cmd, "message-template", "", "{{ .Version }}", "the annotated tag message template. Values: Version,Tag")
	addStringFlag(cmd, "tag-template", "", "{{ .Version }}", "the release tag template. Values: Version")

	return cmd
}

// CreateRelease create project release
func Release(project *project.Project, flags projectReleaseFlags) {

	tagData := struct {
		Version string
	}{
		Version: project.ReleaseVersion(),
	}
	tag := tools.Template(tagData, flags.TagTemplate)

	messageData := struct {
		Version string
		Tag     string
	}{
		Version: tagData.Version,
		Tag:     tag,
	}
	msg := tools.Template(messageData, flags.MessageTemplate)
	tools.Git("tag", "-a", tag, "-m", msg)

	// push project to remote repository
	if flags.Project.SkipPush {
		log.WithField("tag", tag).Info("Skip git push for project release")
	} else {
		tools.Git("push", "--tags")
	}
	log.WithField("version", tag).Info("New release created.")
}
