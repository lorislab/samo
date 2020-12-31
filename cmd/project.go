package cmd

import (
	"github.com/lorislab/samo/project"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type commonFlags struct {
	File              string       `mapstructure:"file"`
	Type              project.Type `mapstructure:"type"`
	Versions          []string     `mapstructure:"version"`
	HashLength        int          `mapstructure:"build-hash"`
	BuildNumberPrefix string       `mapstructure:"build-prefix"`
	BuildNumberLength int          `mapstructure:"build-length"`
}

type projectFlags struct {
	Project           commonFlags `mapstructure:",squash"`
	ReleaseTagMessage string      `mapstructure:"release-tag-message"`
	ReleaseMajor      bool        `mapstructure:"release-major"`
	NextDev           bool        `mapstructure:"next-dev"`
	NextDevMsg        string      `mapstructure:"dev-message"`
	SkipPush          bool        `mapstructure:"skip-push"`
	PatchBranchPrefix string      `mapstructure:"patch-branch"`
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
				Project:   p,
				TagMsg:    op.ReleaseTagMessage,
				Major:     op.ReleaseMajor,
				SkipPush:  op.SkipPush,
				NextDev:   op.NextDev,
				CommitMsg: op.NextDevMsg,
				Versions:  project.CreateVersions(p, []string{"version", "release"}, op.Project.HashLength, op.Project.BuildNumberLength, op.Project.BuildNumberPrefix),
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
				Project:          p,
				Tag:              op.PatchTag,
				PathBranchPrefix: op.PatchBranchPrefix,
				SkipPush:         op.SkipPush,
				NextDev:          op.NextDev,
				CommitMsg:        op.NextDevMsg,
				Versions:         project.CreateVersions(p, nil, 0, 0, ""),
			}
			r.Patch()
		},
		TraverseChildren: true,
	}
)

func init() {
	addChildCmd(rootCmd, projectCmd)
	addSliceFlag(projectCmd, "version", "", []string{project.VerVersion}, "project version type, custom or "+verList)
	addFlag(projectCmd, "build-prefix", "b", "rc", "the build number prefix")
	addIntFlag(projectCmd, "build-length", "e", 3, "the build number length")
	addIntFlag(projectCmd, "build-hash", "", 12, "the git hash length")

	addChildCmd(projectCmd, createReleaseCmd)
	addBoolFlag(createReleaseCmd, "release-major", "", false, "create a major release")
	addFlag(createReleaseCmd, "release-tag-message", "", "{{ .Tag }}", "the release tag message. (template)")
	nd := addBoolFlag(createReleaseCmd, "next-dev", "", true, "update project file (if exists) to next dev version")
	ndm := addFlag(createReleaseCmd, "next-dev-msg", "", "Create new development version {{ .Version }}", "commit message for new development version (template)")
	gsk := addBoolFlag(createReleaseCmd, "skip-push", "", false, "skip git push changes")

	addChildCmd(projectCmd, createPatchCmd)
	addFlagRequired(createPatchCmd, "patch-tag", "", "", "the tag version of the patch branch")
	addFlag(createPatchCmd, "patch-branch", "", "{{ .Major }}.{{ .Minor }}", "patch branch name (template)")
	addFlagRef(createPatchCmd, nd)
	addFlagRef(createPatchCmd, ndm)
	addFlagRef(createPatchCmd, gsk)
}

func readProjectOptions() (projectFlags, project.Project) {
	options := projectFlags{}
	err := viper.Unmarshal(&options)
	if err != nil {
		panic(err)
	}
	log.WithField("options", options).Debug("Load project options")
	return options, loadProject(options.Project.File, project.Type(options.Project.Type))
}
