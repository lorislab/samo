package cmd

import (
	"github.com/lorislab/samo/log"
	"github.com/lorislab/samo/tools"
	"github.com/spf13/cobra"
)

type projectPatchFlags struct {
	Project projectFlags `mapstructure:",squash"`
	Tag     string       `mapstructure:"tag"`
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

	addStringFlagReq(cmd, "tag", "", "", "create patch branch from the release version")

	return cmd
}

// CreatePatch create patch fo the project
func patch(project *Project, flags projectPatchFlags) {

	tagVer := tools.CreateSemVer(flags.Tag)
	if tagVer.Patch() != 0 || len(tagVer.Prerelease()) > 0 {
		log.Fatal("Can not created patch-branch from the patch tag!", log.F("tag", tagVer.Original()))
	}

	branch := createPatchBranchName(tagVer, flags.Project)
	tools.Git("checkout", "-b", branch, flags.Tag)
	log.Debug("Patch branch created", log.F("branch", branch))

	// push changes
	if flags.Project.SkipPush {
		log.Info("Skip git push patch branch", log.F("branch", branch))
	} else {
		tools.Git("push", "-u", "origin", branch)
	}
	log.Info("New patch branch created.", log.F("branch", branch))
}
