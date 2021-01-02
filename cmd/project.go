package cmd

import (
	"github.com/lorislab/samo/project"
	"github.com/spf13/cobra"
)

type commonFlags struct {
	File              string       `mapstructure:"file"`
	Type              project.Type `mapstructure:"type"`
	Versions          []string     `mapstructure:"version"`
	HashLength        int          `mapstructure:"build-hash"`
	BuildNumber       string       `mapstructure:"build-number"`
	BuildNumberLength int          `mapstructure:"build-length"`
	FirstVersion      string       `mapstructure:"first-version"`
}

type projectFlags struct {
	Project           commonFlags `mapstructure:",squash"`
	ReleaseTagMessage string      `mapstructure:"release-tag-message"`
	ReleaseMajor      bool        `mapstructure:"release-major"`
	SkipNextDev       bool        `mapstructure:"skip-next-dev"`
	NextDevMsg        string      `mapstructure:"next-dev-message"`
	SkipPush          bool        `mapstructure:"skip-push"`
	PatchBranch       string      `mapstructure:"patch-branch"`
	PatchTag          string      `mapstructure:"patch-tag"`
}

var (
	verList = project.VersionsText()

	projectCmd = &cobra.Command{
		Use:              "project",
		Short:            "Project operation",
		Long:             `Tasks for the project`,
		TraverseChildren: true,
	}
	createReleaseCmd = &cobra.Command{
		Use:   "release",
		Short: "Create release of the current project and state",
		Long:  `Create release of the current project and state`,
		Run: func(cmd *cobra.Command, args []string) {
			op, p := readProjectOptions()
			r := project.ProjectRequest{
				Project:     p,
				TagMsg:      op.ReleaseTagMessage,
				Major:       op.ReleaseMajor,
				SkipPush:    op.SkipPush,
				SkipNextDev: op.SkipNextDev,
				CommitMsg:   op.NextDevMsg,
				Versions:    createVersionsFrom(p, op.Project, []string{project.VerVersion, project.VerRelease}),
			}
			r.Release()
		},
		TraverseChildren: true,
	}
	createPatchCmd = &cobra.Command{
		Use:   "patch",
		Short: "Create patch of the project release",
		Long:  `Create patch of the project release`,
		Run: func(cmd *cobra.Command, args []string) {
			op, p := readProjectOptions()
			r := project.ProjectRequest{
				Project:     p,
				Tag:         op.PatchTag,
				PathBranch:  op.PatchBranch,
				SkipPush:    op.SkipPush,
				SkipNextDev: op.SkipNextDev,
				CommitMsg:   op.NextDevMsg,
				Versions:    createVersions(p, op.Project),
			}
			r.Patch()
		},
		TraverseChildren: true,
	}
)

func initProject() {
	addChildCmd(rootCmd, projectCmd)
	addSliceFlag(projectCmd, "version", "", []string{project.VerVersion}, "project version type, custom or "+verList)
	addFlag(projectCmd, "build-number", "b", "rc{{ .Number }}.{{ .Hash }}", "the build number (temmplate) [Number,Hash,Count]")
	addIntFlag(projectCmd, "build-length", "e", 3, "the build number length.")
	addIntFlag(projectCmd, "build-hash", "", 12, "the git hash length")
	addFlag(projectCmd, "first-version", "", "0.0.0", "the first version of the project")

	addChildCmd(projectCmd, createReleaseCmd)
	addBoolFlag(createReleaseCmd, "release-major", "", false, "create a major release")
	addFlag(createReleaseCmd, "release-tag-message", "", "{{ .Version }}", "the release tag message. (template) [Version]")
	nd := addBoolFlag(createReleaseCmd, "skip-next-dev", "", false, "skip update project file (if exists) to next dev version")
	ndm := addFlag(createReleaseCmd, "next-dev-message", "", "Create new development version [{{ .Version }}]", "commit message for new development version (template) [Version]")
	skipPush := addBoolFlag(createReleaseCmd, "skip-push", "", false, "skip git push changes")

	addChildCmd(projectCmd, createPatchCmd)
	addFlagReq(createPatchCmd, "patch-tag", "", "", "create patch branch for the release tag")
	addFlag(createPatchCmd, "patch-branch", "", "{{ .Major }}.{{ .Minor }}", "patch branch name (template) [Major,Minor,Patch]")
	addFlagRef(createPatchCmd, nd)
	addFlagRef(createPatchCmd, ndm)
	addFlagRef(createPatchCmd, skipPush)
}

func readProjectOptions() (projectFlags, project.Project) {
	options := projectFlags{}
	readOptions(&options)
	return options, loadProject(options.Project)
}
