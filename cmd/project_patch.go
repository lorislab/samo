package cmd

import (
	"github.com/lorislab/samo/project"
	"github.com/lorislab/samo/tools"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type projectPatchFlags struct {
	Project        projectFlags `mapstructure:",squash"`
	Tag            string       `mapstructure:"tag"`
	BranchTemplate string       `mapstructure:"branch-template"`
}

func createProjectPatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "patch",
		Short: "Create patch of the project release",
		Long:  `Create patch of the project release`,
		Run: func(cmd *cobra.Command, args []string) {
			flags := projectPatchFlags{}
			readOptions(&flags)
			project := loadProject(flags.Project)

			patch(project, flags)
		},
		TraverseChildren: true,
	}

	addStringFlagReq(cmd, "tag", "", "", "create patch branch for the release tag")
	addStringFlag(cmd, "branch-template", "", "{{ .Major }}.{{ .Minor }}", "patch-branch name template. Values: Major,Minor,Patch")

	return cmd
}

// CreatePatch create patch fo the project
func patch(project *project.Project, flags projectPatchFlags) {

	tagVer := tools.CreateSemVer(flags.Tag)
	if tagVer.Patch() != 0 || len(tagVer.Prerelease()) > 0 {
		log.WithField("tag", tagVer.Original()).Fatal("Can not created patch-branch from the patch tag!")
	}

	branch := tools.Template(tagVer, flags.BranchTemplate)
	tools.Git("checkout", "-b", branch, flags.Tag)
	log.WithField("branch", branch).Debug("Patch branch created")

	// push changes
	if flags.Project.SkipPush {
		log.WithField("branch", branch).Info("Skip git push patch branch")
	} else {
		tools.Git("push", "-u", "origin", branch)
	}
	log.WithField("branch", branch).Info("New patch branch created.")
}
